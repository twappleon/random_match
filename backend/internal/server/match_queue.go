package server

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"

	"random-match/backend/internal/model"
)

// Atomic join: enqueue each user at most once. Premium users wait in a separate
// queue, and every pairing consumes premium waiters before regular waiters.
var matchJoinScript = redis.NewScript(`
if redis.call('SISMEMBER', KEYS[3], ARGV[1]) == 1 then
  return {2}
end
redis.call('SADD', KEYS[3], ARGV[1])
if ARGV[2] == 'premium' then
  redis.call('RPUSH', KEYS[1], ARGV[1])
else
  redis.call('RPUSH', KEYS[2], ARGV[1])
end
local len = redis.call('LLEN', KEYS[1]) + redis.call('LLEN', KEYS[2])
if len < 2 then
  return {0}
end
local u1 = redis.call('LPOP', KEYS[1])
if not u1 then
  u1 = redis.call('LPOP', KEYS[2])
end
local u2 = redis.call('LPOP', KEYS[1])
if not u2 then
  u2 = redis.call('LPOP', KEYS[2])
end
redis.call('SREM', KEYS[3], u1, u2)
if u1 == ARGV[1] then
  return {1, u2}
end
return {1, u1}
`)

type matchJoinResult struct {
	waiting bool
	partner model.MatchTicket
}

func enqueueAndMatch(ctx context.Context, client *redis.Client, queueKey string, ticket model.MatchTicket, priority bool) (matchJoinResult, error) {
	priorityKey := queueKey + ":premium"
	regularKey := queueKey + ":regular"
	membersKey := queueKey + ":members"
	tier := "regular"
	if priority {
		tier = "premium"
	}
	raw, err := matchJoinScript.Run(ctx, client, []string{priorityKey, regularKey, membersKey}, ticket.UserID, tier).Result()
	if err != nil {
		return matchJoinResult{}, err
	}

	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return matchJoinResult{}, errors.New("unexpected match script result")
	}

	status, ok := items[0].(int64)
	if !ok {
		return matchJoinResult{}, errors.New("unexpected match script status")
	}
	if status == 0 || status == 2 {
		return matchJoinResult{waiting: true}, nil
	}
	if len(items) != 2 {
		return matchJoinResult{}, errors.New("unexpected match script payload")
	}

	partnerID, ok := items[1].(string)
	if !ok || partnerID == "" || partnerID == ticket.UserID {
		return matchJoinResult{waiting: true}, nil
	}

	partner := model.MatchTicket{
		UserID: partnerID,
		Mode:   ticket.Mode,
		Region: ticket.Region,
	}
	return matchJoinResult{partner: partner}, nil
}

func queueForPriority(queueKey string, priority bool) string {
	if priority {
		return queueKey + ":premium"
	}
	return queueKey + ":regular"
}

func removeUserFromMatchQueues(ctx context.Context, client *redis.Client, userID string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, "match:queue:v2:*", 100).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			if strings.HasSuffix(key, ":members") {
				continue
			}
			if strings.HasSuffix(key, ":regular") || strings.HasSuffix(key, ":premium") {
				if err := client.LRem(ctx, key, 0, userID).Err(); err != nil {
					return err
				}
				if err := client.SRem(ctx, strings.TrimSuffix(strings.TrimSuffix(key, ":regular"), ":premium")+":members", userID).Err(); err != nil {
					return err
				}
				continue
			}
			if err := client.LRem(ctx, key, 0, userID).Err(); err != nil {
				return err
			}
			if err := client.SRem(ctx, key+":members", userID).Err(); err != nil {
				return err
			}
		}
		if nextCursor == 0 {
			return nil
		}
		cursor = nextCursor
	}
}

func countWaitingUsers(ctx context.Context, client *redis.Client) (int, error) {
	var cursor uint64
	total := 0
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, "match:queue:v2:*", 100).Result()
		if err != nil {
			return 0, err
		}
		for _, key := range keys {
			if strings.HasSuffix(key, ":members") {
				continue
			}
			count, err := client.LLen(ctx, key).Result()
			if err != nil {
				return 0, err
			}
			total += int(count)
		}
		if nextCursor == 0 {
			return total, nil
		}
		cursor = nextCursor
	}
}
