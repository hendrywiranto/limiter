package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/hendrywiranto/limiter"
	"github.com/hendrywiranto/limiter/redis"
)

type RedisSuite struct {
	suite.Suite

	ctx       context.Context
	ctrl      *gomock.Controller
	redisMock redismock.ClientMock
	adapter   *redis.Adapter
}

func TestCalculateReceiptTransactionFeeDetail(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())

	client, mock := redismock.NewClientMock()
	s.redisMock = mock
	s.adapter = redis.NewAdapter(client)
}

// ==================== Get Cases ====================

func (s *RedisSuite) TestGetKeyFound() {
	s.redisMock.ExpectGet("mykey").SetVal("16")

	var value int64
	err := s.adapter.Get(s.ctx, "mykey", &value)

	s.NoError(err)
	s.Equal(int64(16), value)
}

func (s *RedisSuite) TestGetKeyNotFound() {
	s.redisMock.ExpectGet("mykey").RedisNil()

	var value int64
	err := s.adapter.Get(s.ctx, "mykey", &value)

	s.Error(err)
	s.ErrorIs(err, limiter.ErrCacheMiss)
}

func (s *RedisSuite) TestGetError() {
	s.redisMock.ExpectGet("mykey").SetErr(errors.New("some error"))

	var value int64
	err := s.adapter.Get(s.ctx, "mykey", &value)

	s.Error(err)
	s.ErrorContains(err, "some error")
}

// ==================== Set Cases ====================

func (s *RedisSuite) TestSet() {
	s.redisMock.ExpectSet("mykey", int64(16), time.Hour).SetVal("16")

	err := s.adapter.Set(s.ctx, "mykey", 16, time.Hour)

	s.NoError(err)
}

func (s *RedisSuite) TestSetError() {
	s.redisMock.ExpectSet("mykey", int64(16), time.Hour).SetErr(errors.New("some error"))

	err := s.adapter.Set(s.ctx, "mykey", 16, time.Hour)

	s.Error(err)
	s.ErrorContains(err, "some error")
}

// ==================== IncrBy Cases ====================

func (s *RedisSuite) TestIncrBy() {
	s.redisMock.ExpectIncrBy("mykey", int64(16)).SetVal(16)

	err := s.adapter.IncrBy(s.ctx, "mykey", 16)

	s.NoError(err)
}

func (s *RedisSuite) TestIncrByError() {
	s.redisMock.ExpectIncrBy("mykey", int64(16)).SetErr(errors.New("some error"))

	err := s.adapter.IncrBy(s.ctx, "mykey", 16)

	s.Error(err)
	s.ErrorContains(err, "some error")
}

// ==================== SumKeys Cases ====================

func (s *RedisSuite) TestSumKeys() {
	s.redisMock.ExpectMGet("key1", "key2").SetVal([]interface{}{int64(1), int64(2)})

	sum, err := s.adapter.SumKeys(s.ctx, []string{"key1", "key2"})

	s.NoError(err)
	s.Equal(int64(3), sum)
}

func (s *RedisSuite) TestSumKeysError() {
	s.redisMock.ExpectMGet("key1", "key2").SetErr(errors.New("some error"))

	sum, err := s.adapter.SumKeys(s.ctx, []string{"key1", "key2"})

	s.Error(err)
	s.ErrorContains(err, "some error")
	s.Empty(sum)
}
