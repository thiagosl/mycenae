package plot

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gocql/gocql"
	"github.com/uol/gobol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (persist *persistence) GetTST(keyspace string, keys []string, start, end int64, search *regexp.Regexp, allowFullFetch bool, maxBytesLimit uint32, keyset string) (map[string][]TextPnt, uint32, gobol.Error) {

	track := time.Now()
	start--
	end++

	var tsid string
	var date int64
	var value string
	var err error
	var numBytes uint32
	idsGroup := persist.buildInGroup(keys)

	iter := persist.cassandra.Query(
		fmt.Sprintf(
			`SELECT id, date, value FROM %v.ts_text_stamp WHERE id in (%s) AND date > ? AND date < ? ALLOW FILTERING`,
			keyspace,
			idsGroup,
		),
		start,
		end,
	).Iter()

	tsMap := map[string][]TextPnt{}
	countRows := 0
	limitReached := false

	for iter.Scan(&tsid, &date, &value) {
		add := true

		if search != nil && !search.MatchString(value) {
			add = false
		}

		if add {
			point := TextPnt{
				Date:  date,
				Value: value,
			}

			if _, ok := tsMap[tsid]; !ok {
				numBytes += uint32(persist.getStringSize(tsid))
			}

			tsMap[tsid] = append(tsMap[tsid], point)

			numBytes += uint32(persist.constPartBytesFromTextPoint + persist.getStringSize(value))

			if !allowFullFetch && numBytes >= maxBytesLimit {
				limitReached = true
				break
			}

			countRows++
		}
	}

	statsValueAdd(
		"scylla.query.bytes",
		map[string]string{
			"keyset":   keyset,
			"keyspace": keyspace,
			"type":     "number",
		},
		float64(numBytes),
	)

	if err = iter.Close(); err != nil {
		fields := []zapcore.Field{
			zap.String("package", "plot/persistence"),
			zap.String("func", "getTST"),
		}
		gblog.Error(err.Error(), fields...)

		if err == gocql.ErrNotFound {
			return map[string][]TextPnt{}, 0, errNoContent("getTST")
		}

		statsSelectFerror(keyspace, "ts_text_stamp")
		return map[string][]TextPnt{}, 0, errPersist("getTST", err)
	}

	statsSelect(keyspace, "ts_text_stamp", time.Since(track), countRows)

	if limitReached && !allowFullFetch {
		return map[string][]TextPnt{}, numBytes, errMaxBytesLimitWrapper("GetTS", persist.maxBytesErr)
	}

	return tsMap, numBytes, nil
}
