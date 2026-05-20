package server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"

	"random-match/backend/internal/model"
)

// Atomic join: RPUSH self, then if queue length >= 2, pop two and pair.
var matchJoinScript = redis.NewScript(`
redis.call('RPUSH', KEYS[1], ARGV[1])
local len = redis.call('LLEN', KEYS[1])
if len < 2 then
  return {0}
end
local t1 = redis.call('LPOP', KEYS[1])
local t2 = redis.call('LPOP', KEYS[1])
return {1, t1, t2}
`)

type matchJoinResult struct {
	waiting bool
	partner model.MatchTicket
}

func enqueueAndMatch(ctx context.Context, client *redis.Client, queueKey string, ticket model.MatchTicket) (matchJoinResult, error) {
	payload, err := json.Marshal(ticket)
	if err != nil {
		return matchJoinResult{}, err
	}

	raw, err := matchJoinScript.Run(ctx, client, []string{queueKey}, payload).Result()
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
	if len(items) != 3 {
		return matchJoinResult{}, errors.New("unexpected match script payload")
	}

	var first, second model.MatchTicket
	if err := json.Unmarshal([]byte(items[1].(string)), &first); err != nil {
		return matchJoinResult{}, err
	}
	if err := json.Unmarshal([]byte(items[2].(string)), &second); err != nil {
		return matchJoinResult{}, err
	}

	if first.UserID == second.UserID {
		// Duplicate tickets for the same user; keep one and wait.
		one, _ := json.Marshal(first)
		if err := client.RPush(ctx, queueKey, one).Err(); err != nil {
			return matchJoinResult{}, err
		}
		return matchJoinResult{waiting: true}, nil
	}

	partner := first
	if first.UserID == ticket.UserID {
		partner = second
	} else if second.UserID != ticket.UserID {
		// Neither popped ticket is the current user (should not happen); re-queue both.
		p1, _ := json.Marshal(first)
		p2, _ := json.Marshal(second)
		_ = client.RPush(ctx, queueKey, p1, p2).Err()
		return matchJoinResult{waiting: true}, nil
	}

	if partner.UserID == ticket.UserID {
		one, _ := json.Marshal(first)
		if err := client.RPush(ctx, queueKey, one).Err(); err != nil {
			return matchJoinResult{}, err
		}
		return matchJoinResult{waiting: true}, nil
	}

	return matchJoinResult{partner: partner}, nil
}
