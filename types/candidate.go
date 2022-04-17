// package types

// import (
// 	"context"
// 	"errors"
// 	"time"

// 	"github.com/go-redis/redis"
// )

// type Candidate struct {
// 	Address string
// }

// func NewCandidate(address string) Candidate {
// 	return Candidate{
// 		Address: address,
// 	}
// }
// func (c Candidate) Update(ctx context.Context) error {
// 	timestamp := time.Now().UTC().Unix()
// 	get := rdb.Get(ctx, "peer:by:"+c.Address)
// 	if err := get.Err(); err == nil {
// 		return errors.New("already exists as a peer")
// 	}

// 	intCmd := rdb.ZAdd(ctx, "candidates", &redis.Z{Score: float64(timestamp), Member: c.Address})
// 	return intCmd.Err()
// }
// func (c Candidate) Delete(ctx context.Context) error {
// 	intCmd := rdb.ZRem(ctx, "candidates", c.Address)

// 	return intCmd.Err()
// }
