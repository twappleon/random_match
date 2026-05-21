package server

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"random-match/backend/internal/model"
)

// Atomic join: enqueue each user at most once, then pair the first two users.
var matchJoinScript = redis.NewScript(`
if redis.call('SISMEMBER', KEYS[2], ARGV[1]) == 1 then
  return {0}
end
redis.call('SADD', KEYS[2], ARGV[1])
redis.call('RPUSH', KEYS[1], ARGV[1])
local len = redis.call('LLEN', KEYS[1])
if len < 2 then
  return {0}
end
local u1 = redis.call('LPOP', KEYS[1])
local u2 = redis.call('LPOP', KEYS[1])
redis.call('SREM', KEYS[2], u1, u2)
if u1 == ARGV[1] then
  return {1, u2}
end
return {1, u1}
`)

type matchJoinResult struct {
	waiting bool
	partner model.MatchTicket
}

func enqueueAndMatch(ctx context.Context, client *redis.Client, queueKey string, ticket model.MatchTicket) (matchJoinResult, error) {
	membersKey := queueKey + ":members"
	raw, err := matchJoinScript.Run(ctx, client, []string{queueKey, membersKey}, ticket.UserID).Result()
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
	if status == 0 {
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
