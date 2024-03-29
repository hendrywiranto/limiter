package limiter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hendrywiranto/limiter"
	"github.com/hendrywiranto/limiter/mock"
	"github.com/stretchr/testify/suite"
)

type LimiterSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	ctx  context.Context

	adapter *mock.MockAdapter
	l       *limiter.Limiter
}

func (s *LimiterSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())

	s.adapter = mock.NewMockAdapter(s.ctrl)
	limits := map[string]limiter.Limits{
		"metric_test": {
			limiter.DurationDay:    300,
			limiter.DurationHour:   30,
			limiter.DurationMinute: 10,
			limiter.DurationSecond: 5,
		},
	}
	s.l = limiter.New(s.adapter, limits)

	// mock the current time to 2024-02-29 23:11:11 UTC.
	limiter.Now = func() time.Time {
		return time.Date(2024, 0o2, 29, 23, 11, 11, 0, time.UTC)
	}
}

func TestLimiter(t *testing.T) {
	suite.Run(t, new(LimiterSuite))
}

func (s *LimiterSuite) TestRecordAllSuccess() {
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:20240229231111", int64(10)).Return(nil)
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:202402292311", int64(10)).Return(nil)
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:2024022923", int64(10)).Return(nil)

	err := s.l.Record(s.ctx, "metric_test", 10)
	s.NoError(err)
}

func (s *LimiterSuite) TestRecordMetricNotFound() {
	err := s.l.Record(s.ctx, "unknown_metric", 10)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrMetricNotFound)
}

func (s *LimiterSuite) TestRecordFailedSecond() {
	mockedErr := errors.New("mocked error")
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:20240229231111", int64(10)).Return(mockedErr)

	err := s.l.Record(s.ctx, "metric_test", 10)
	s.Error(err)
	s.ErrorIs(err, mockedErr)
}

func (s *LimiterSuite) TestRecordFailedMinute() {
	mockedErr := errors.New("mocked error")
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:20240229231111", int64(10)).Return(nil)
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:202402292311", int64(10)).Return(mockedErr)

	err := s.l.Record(s.ctx, "metric_test", 10)
	s.Error(err)
	s.ErrorIs(err, mockedErr)
}

func (s *LimiterSuite) TestRecordFailedHour() {
	mockedErr := errors.New("mocked error")
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:20240229231111", int64(10)).Return(nil)
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:202402292311", int64(10)).Return(nil)
	s.adapter.EXPECT().IncrBy(s.ctx, "metric_test:2024022923", int64(10)).Return(mockedErr)

	err := s.l.Record(s.ctx, "metric_test", 10)
	s.Error(err)
	s.ErrorIs(err, mockedErr)
}

func (s *LimiterSuite) TestCheckMetricNotFound() {
	err := s.l.Check(s.ctx, "unknown_metric", limiter.DurationDay)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrMetricNotFound)
}

func (s *LimiterSuite) TestCheckDayLimitNotSet() {
	limits := map[string]limiter.Limits{
		"metric_test": {},
	}
	s.l = limiter.New(s.adapter, limits)
	s.adapter.EXPECT().SumKeys(s.ctx, dayKeys).Return(int64(250), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationDay)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrLimitNotSet)
}

func (s *LimiterSuite) TestCheckHourLimitNotSet() {
	limits := map[string]limiter.Limits{
		"metric_test": {},
	}
	s.l = limiter.New(s.adapter, limits)
	s.adapter.EXPECT().SumKeys(s.ctx, hourKeys).Return(int64(25), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationHour)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrLimitNotSet)
}

func (s *LimiterSuite) TestCheckMinuteLimitNotSet() {
	limits := map[string]limiter.Limits{
		"metric_test": {},
	}
	s.l = limiter.New(s.adapter, limits)
	s.adapter.EXPECT().SumKeys(s.ctx, minuteKeys).Return(int64(5), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationMinute)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrLimitNotSet)
}

func (s *LimiterSuite) TestCheckSecondLimitNotSet() {
	limits := map[string]limiter.Limits{
		"metric_test": {},
	}
	s.l = limiter.New(s.adapter, limits)
	s.adapter.EXPECT().SumKeys(s.ctx, []string{"20240229231110"}).Return(int64(2), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationSecond)
	s.Error(err)
	s.ErrorIs(err, limiter.ErrLimitNotSet)
}

func (s *LimiterSuite) TestCheckDayWithinLimit() {
	s.adapter.EXPECT().SumKeys(s.ctx, dayKeys).Return(int64(250), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationDay)
	s.NoError(err)
}

func (s *LimiterSuite) TestCheckHourWithinLimit() {
	s.adapter.EXPECT().SumKeys(s.ctx, hourKeys).Return(int64(25), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationHour)
	s.NoError(err)
}

func (s *LimiterSuite) TestCheckMinuteWithinLimit() {
	s.adapter.EXPECT().SumKeys(s.ctx, minuteKeys).Return(int64(5), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationMinute)
	s.NoError(err)
}

func (s *LimiterSuite) TestCheckSecondWithinLimit() {
	s.adapter.EXPECT().SumKeys(s.ctx, []string{"20240229231110"}).Return(int64(2), nil)

	err := s.l.Check(s.ctx, "metric_test", limiter.DurationSecond)
	s.NoError(err)
}

func (s *LimiterSuite) TestGenerateKeysDay() {
	keys := s.l.GenerateKeys(limiter.DurationDay)
	s.Len(keys, 142)
	s.Equal(dayKeys, keys)
}

func (s *LimiterSuite) TestGenerateKeysHour() {
	keys := s.l.GenerateKeys(limiter.DurationHour)
	s.Len(keys, 119)
	s.Equal(hourKeys, keys)
}

func (s *LimiterSuite) TestGenerateKeysMinute() {
	keys := s.l.GenerateKeys(limiter.DurationMinute)
	s.Len(keys, 60)
	s.Equal(minuteKeys, keys)
}

func (s *LimiterSuite) TestGenerateKeysSecond() {
	keys := s.l.GenerateKeys(limiter.DurationSecond)
	s.Len(keys, 1)
	s.Equal("20240229231110", keys[0])
}

var (
	minuteKeys = []string{
		"20240229231011",
		"20240229231012",
		"20240229231013",
		"20240229231014",
		"20240229231015",
		"20240229231016",
		"20240229231017",
		"20240229231018",
		"20240229231019",
		"20240229231020",
		"20240229231021",
		"20240229231022",
		"20240229231023",
		"20240229231024",
		"20240229231025",
		"20240229231026",
		"20240229231027",
		"20240229231028",
		"20240229231029",
		"20240229231030",
		"20240229231031",
		"20240229231032",
		"20240229231033",
		"20240229231034",
		"20240229231035",
		"20240229231036",
		"20240229231037",
		"20240229231038",
		"20240229231039",
		"20240229231040",
		"20240229231041",
		"20240229231042",
		"20240229231043",
		"20240229231044",
		"20240229231045",
		"20240229231046",
		"20240229231047",
		"20240229231048",
		"20240229231049",
		"20240229231050",
		"20240229231051",
		"20240229231052",
		"20240229231053",
		"20240229231054",
		"20240229231055",
		"20240229231056",
		"20240229231057",
		"20240229231058",
		"20240229231059",
		"20240229231100",
		"20240229231101",
		"20240229231102",
		"20240229231103",
		"20240229231104",
		"20240229231105",
		"20240229231106",
		"20240229231107",
		"20240229231108",
		"20240229231109",
		"20240229231110",
	}

	hourKeys = []string{
		"20240229221111",
		"20240229221112",
		"20240229221113",
		"20240229221114",
		"20240229221115",
		"20240229221116",
		"20240229221117",
		"20240229221118",
		"20240229221119",
		"20240229221120",
		"20240229221121",
		"20240229221122",
		"20240229221123",
		"20240229221124",
		"20240229221125",
		"20240229221126",
		"20240229221127",
		"20240229221128",
		"20240229221129",
		"20240229221130",
		"20240229221131",
		"20240229221132",
		"20240229221133",
		"20240229221134",
		"20240229221135",
		"20240229221136",
		"20240229221137",
		"20240229221138",
		"20240229221139",
		"20240229221140",
		"20240229221141",
		"20240229221142",
		"20240229221143",
		"20240229221144",
		"20240229221145",
		"20240229221146",
		"20240229221147",
		"20240229221148",
		"20240229221149",
		"20240229221150",
		"20240229221151",
		"20240229221152",
		"20240229221153",
		"20240229221154",
		"20240229221155",
		"20240229221156",
		"20240229221157",
		"20240229221158",
		"20240229221159",
		"202402292212",
		"202402292213",
		"202402292214",
		"202402292215",
		"202402292216",
		"202402292217",
		"202402292218",
		"202402292219",
		"202402292220",
		"202402292221",
		"202402292222",
		"202402292223",
		"202402292224",
		"202402292225",
		"202402292226",
		"202402292227",
		"202402292228",
		"202402292229",
		"202402292230",
		"202402292231",
		"202402292232",
		"202402292233",
		"202402292234",
		"202402292235",
		"202402292236",
		"202402292237",
		"202402292238",
		"202402292239",
		"202402292240",
		"202402292241",
		"202402292242",
		"202402292243",
		"202402292244",
		"202402292245",
		"202402292246",
		"202402292247",
		"202402292248",
		"202402292249",
		"202402292250",
		"202402292251",
		"202402292252",
		"202402292253",
		"202402292254",
		"202402292255",
		"202402292256",
		"202402292257",
		"202402292258",
		"202402292259",
		"202402292300",
		"202402292301",
		"202402292302",
		"202402292303",
		"202402292304",
		"202402292305",
		"202402292306",
		"202402292307",
		"202402292308",
		"202402292309",
		"202402292310",
		"20240229231100",
		"20240229231101",
		"20240229231102",
		"20240229231103",
		"20240229231104",
		"20240229231105",
		"20240229231106",
		"20240229231107",
		"20240229231108",
		"20240229231109",
		"20240229231110",
	}

	dayKeys = []string{
		"20240228231111",
		"20240228231112",
		"20240228231113",
		"20240228231114",
		"20240228231115",
		"20240228231116",
		"20240228231117",
		"20240228231118",
		"20240228231119",
		"20240228231120",
		"20240228231121",
		"20240228231122",
		"20240228231123",
		"20240228231124",
		"20240228231125",
		"20240228231126",
		"20240228231127",
		"20240228231128",
		"20240228231129",
		"20240228231130",
		"20240228231131",
		"20240228231132",
		"20240228231133",
		"20240228231134",
		"20240228231135",
		"20240228231136",
		"20240228231137",
		"20240228231138",
		"20240228231139",
		"20240228231140",
		"20240228231141",
		"20240228231142",
		"20240228231143",
		"20240228231144",
		"20240228231145",
		"20240228231146",
		"20240228231147",
		"20240228231148",
		"20240228231149",
		"20240228231150",
		"20240228231151",
		"20240228231152",
		"20240228231153",
		"20240228231154",
		"20240228231155",
		"20240228231156",
		"20240228231157",
		"20240228231158",
		"20240228231159",
		"202402282312",
		"202402282313",
		"202402282314",
		"202402282315",
		"202402282316",
		"202402282317",
		"202402282318",
		"202402282319",
		"202402282320",
		"202402282321",
		"202402282322",
		"202402282323",
		"202402282324",
		"202402282325",
		"202402282326",
		"202402282327",
		"202402282328",
		"202402282329",
		"202402282330",
		"202402282331",
		"202402282332",
		"202402282333",
		"202402282334",
		"202402282335",
		"202402282336",
		"202402282337",
		"202402282338",
		"202402282339",
		"202402282340",
		"202402282341",
		"202402282342",
		"202402282343",
		"202402282344",
		"202402282345",
		"202402282346",
		"202402282347",
		"202402282348",
		"202402282349",
		"202402282350",
		"202402282351",
		"202402282352",
		"202402282353",
		"202402282354",
		"202402282355",
		"202402282356",
		"202402282357",
		"202402282358",
		"202402282359",
		"2024022900",
		"2024022901",
		"2024022902",
		"2024022903",
		"2024022904",
		"2024022905",
		"2024022906",
		"2024022907",
		"2024022908",
		"2024022909",
		"2024022910",
		"2024022911",
		"2024022912",
		"2024022913",
		"2024022914",
		"2024022915",
		"2024022916",
		"2024022917",
		"2024022918",
		"2024022919",
		"2024022920",
		"2024022921",
		"2024022922",
		"202402292300",
		"202402292301",
		"202402292302",
		"202402292303",
		"202402292304",
		"202402292305",
		"202402292306",
		"202402292307",
		"202402292308",
		"202402292309",
		"202402292310",
		"20240229231100",
		"20240229231101",
		"20240229231102",
		"20240229231103",
		"20240229231104",
		"20240229231105",
		"20240229231106",
		"20240229231107",
		"20240229231108",
		"20240229231109",
		"20240229231110",
	}
)
