package server

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"random-match/backend/internal/model"
)

type matchJoinResult struct {
	waiting bool
	partner model.MatchTicket
}

type matchCompatibilityFunc func(model.MatchTicket) (bool, error)
type requeueItem struct {
	key string
	id  string
}

func enqueueAndMatch(ctx context.Context, client *redis.Client, queueKey string, ticket model.MatchTicket, priority bool, compatible matchCompatibilityFunc) (matchJoinResult, error) {
	priorityKey := queueKey + ":premium"
	regularKey := queueKey + ":regular"
	membersKey := queueKey + ":members"
	selfQueueKey := queueForPriority(queueKey, priority)

	if err := queueTicket(ctx, client, queueKey, ticket, priority); err != nil {
		return matchJoinResult{}, err
	}

	requeue := make([]requeueItem, 0, 32)
	selfPopped := false
	for _, key := range []string{priorityKey, regularKey} {
		for attempts := 0; attempts < 32; attempts++ {
			candidateID, err := client.LPop(ctx, key).Result()
			if err == redis.Nil {
				break
			}
			if err != nil {
				return matchJoinResult{}, err
			}
			if candidateID == "" {
				continue
			}
			if candidateID == ticket.UserID {
				selfPopped = true
				continue
			}
			candidateTicket, ok := getQueuedTicket(ctx, client, queueKey, candidateID)
			if !ok {
				if err := client.SRem(ctx, membersKey, candidateID).Err(); err != nil {
					return matchJoinResult{}, err
				}
				continue
			}
			matches, err := compatible(candidateTicket)
			if err != nil {
				requeue = append(requeue, requeueItem{key: key, id: candidateID})
				return matchJoinResult{}, err
			}
			if !matches {
				requeue = append(requeue, requeueItem{key: key, id: candidateID})
				continue
			}
			if err := requeueCandidates(ctx, client, requeue); err != nil {
				return matchJoinResult{}, err
			}
			pipe := client.TxPipeline()
			pipe.SRem(ctx, membersKey, ticket.UserID, candidateID)
			pipe.Del(ctx, queueKey+":ticket:"+ticket.UserID, queueKey+":ticket:"+candidateID)
			pipe.LRem(ctx, priorityKey, 0, ticket.UserID)
			pipe.LRem(ctx, regularKey, 0, ticket.UserID)
			if _, err := pipe.Exec(ctx); err != nil {
				return matchJoinResult{}, err
			}
			return matchJoinResult{partner: candidateTicket}, nil
		}
	}
	if err := requeueCandidates(ctx, client, requeue); err != nil {
		return matchJoinResult{}, err
	}
	if selfPopped {
		if err := client.RPush(ctx, selfQueueKey, ticket.UserID).Err(); err != nil {
			return matchJoinResult{}, err
		}
	}
	return matchJoinResult{waiting: true}, nil
}

func queueTicket(ctx context.Context, client *redis.Client, queueKey string, ticket model.MatchTicket, priority bool) error {
	payload, err := json.Marshal(ticket)
	if err != nil {
		return err
	}
	priorityKey := queueKey + ":premium"
	regularKey := queueKey + ":regular"
	pipe := client.TxPipeline()
	pipe.Set(ctx, queueKey+":ticket:"+ticket.UserID, payload, 15*time.Minute)
	pipe.SAdd(ctx, queueKey+":members", ticket.UserID)
	pipe.LRem(ctx, priorityKey, 0, ticket.UserID)
	pipe.LRem(ctx, regularKey, 0, ticket.UserID)
	pipe.RPush(ctx, queueForPriority(queueKey, priority), ticket.UserID)
	_, err = pipe.Exec(ctx)
	return err
}

func getQueuedTicket(ctx context.Context, client *redis.Client, queueKey, userID string) (model.MatchTicket, bool) {
	raw, err := client.Get(ctx, queueKey+":ticket:"+userID).Result()
	if err != nil {
		return model.MatchTicket{}, false
	}
	var ticket model.MatchTicket
	if err := json.Unmarshal([]byte(raw), &ticket); err != nil || ticket.UserID == "" {
		return model.MatchTicket{}, false
	}
	return ticket, true
}

func requeueCandidates(ctx context.Context, client *redis.Client, items []requeueItem) error {
	if len(items) == 0 {
		return nil
	}
	pipe := client.TxPipeline()
	for _, item := range items {
		pipe.RPush(ctx, item.key, item.id)
	}
	_, err := pipe.Exec(ctx)
	return err
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
		keys, nextCursor, err := client.Scan(ctx, cursor, "match:queue:v*:*", 100).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			if strings.Contains(key, ":ticket:") {
				if strings.HasSuffix(key, ":"+userID) {
					if err := client.Del(ctx, key).Err(); err != nil {
						return err
					}
				}
				continue
			}
			if strings.HasSuffix(key, ":members") {
				if err := client.SRem(ctx, key, userID).Err(); err != nil {
					return err
				}
				continue
			}
			if strings.HasSuffix(key, ":regular") || strings.HasSuffix(key, ":premium") {
				if err := client.LRem(ctx, key, 0, userID).Err(); err != nil {
					return err
				}
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
		keys, nextCursor, err := client.Scan(ctx, cursor, "match:queue:v*:*", 100).Result()
		if err != nil {
			return 0, err
		}
		for _, key := range keys {
			if !strings.HasSuffix(key, ":regular") && !strings.HasSuffix(key, ":premium") {
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
