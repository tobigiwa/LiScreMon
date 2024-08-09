package database

import (
	"cmp"
	"errors"
	"fmt"

	"slices"
	utils "utils"

	badger "github.com/dgraph-io/badger/v4"
)

func (bs *BadgerDBStore) GetDay(date utils.Date) (DailyStat, error) {

	if day := utils.ToTimeType(date); day.After(utils.FormattedToDay()) {
		return ZeroValueDailyStat, ErrFutureDay
	}

	byteData, err := bs.Get(dbDayKey(date))
	errKeyNotFound := errors.Is(err, badger.ErrKeyNotFound)

	if err != nil && !errKeyNotFound {
		return ZeroValueDailyStat, err
	}

	if errKeyNotFound {
		return bs.getDailyAppStat(date)
	}

	dayStat, err := utils.DecodeJSON[DailyStat](byteData)
	if err != nil {
		return ZeroValueDailyStat, err
	}
	return dayStat, nil
}

func (bs *BadgerDBStore) getDailyAppStat(day utils.Date) (DailyStat, error) {
	var (
		result       DailyStat
		dayTotalData utils.Stats
		arr          = make([]utils.AppStat, 0, 20)
	)

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := dbAppPrefix
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(v []byte) error {

				var (
					app            AppInfo
					appStatArrData utils.AppStat
					err            error
				)

				if app, err = utils.DecodeJSON[AppInfo](v); err != nil {
					return err
				}

				thatDayStat, ok := app.ScreenStat[day]
				if ok { // if the app has that day
					// EachApp []appstat
					appStatArrData.AppName = app.AppName
					appStatArrData.Usage.Active = thatDayStat.Active
					appStatArrData.Usage.Inactive = thatDayStat.Inactive
					appStatArrData.Usage.Open = thatDayStat.Open
					arr = append(arr, appStatArrData)

					// DayTotall stats
					dayTotalData.Active += thatDayStat.Active
					dayTotalData.Inactive += thatDayStat.Inactive
					dayTotalData.Open += thatDayStat.Open
				}

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ZeroValueDailyStat, err
	}

	slices.SortFunc(arr, func(a, b utils.AppStat) int {
		return cmp.Compare(b.Usage.Active, a.Usage.Active)
	})

	result.DayTotal.Active = dayTotalData.Active
	result.DayTotal.Inactive = dayTotalData.Inactive
	result.DayTotal.Open = dayTotalData.Open
	result.EachApp = arr

	if day != utils.Today() {
		byteData, _ := utils.EncodeJSON(result)
		if err := bs.setOrUpdateKeyValue(dbDayKey(day), byteData); err != nil {
			fmt.Println("ERROR WRITING NEW DAY ENTRY", day, "ERROR IS:", err)
		}
		fmt.Println("WRITING NEW DAY ENTRY", day)
	}

	return result, nil
}
