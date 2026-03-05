package maria

import (
	"context"
	"fmt"
	"pjt/internal/logger"
	"strings"
	"time"
)

// 확장 INSERT 문
func (i *Maria) InsertEventLog(tableName string, params [][]any) error {
	timeout, cancel := context.WithTimeout(i.ctx, i.timeout)
	defer cancel()

	var query strings.Builder
	query.WriteString(fmt.Sprintf(`INSERT INTO %s (params...) VALUES `, tableName))
	size := len(params)
	for idx, param := range params {
		query.WriteString(fmt.Sprintf(`(%d, %d, %d, str_to_date('%s','%%Y-%%m-%%d %%H:%%i:%%s'),%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f)`, param...))
		if idx == size-1 {
			continue
		}
		query.WriteString(`, `)
	}

	if _, err := i.db.ExecContext(timeout, query.String()); err != nil {
		logger.Infoln(err)
		return err
	}

	return nil
}

// Native Query, Paging
func (i *Maria) SelectEvtLogPaged(tableName string, startDT, endDT time.Time, evtLogReq any) (any, error) {
	timeout, cancel := context.WithTimeout(i.ctx, i.timeout)
	defer cancel()
	page := 1
	size := 100
	order := 1
	// limit에 1을 더하는 이유는 다음 페이지가 존재하는지 확인하기 위함.
	// e.g. pageSize: 100, page: 2 -> startRow: 101, endRow: 201
	offset := (page - 1) * size
	limit := size + 1

	var sortOrder string
	if order == 0 {
		sortOrder = `ASC`
	} else {
		sortOrder = `DESC`
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`
        SELECT event_domain, event_category, event_content, coll_dt, subject_type, subject_id1, subject_id2,
               target_type, target_id1, target_id2, result_code, value_param1, value_param2, value_param3, value_param4, value_param5, value_param6, value_param7
        FROM %s `, tableName))

	// check := false
	// if len(evtLogReq.CodeArr3) != 0 {
	// 	sb.WriteString(fmt.Sprintf(`WHERE event_domain = %d AND event_category = %d AND event_content in( `, evtLogReq.Code1, evtLogReq.Code2))
	// 	for idx, code3 := range evtLogReq.CodeArr3 {
	// 		sb.WriteString(fmt.Sprintf(`%d`, code3))
	// 		if len(evtLogReq.CodeArr3)-1 != idx {
	// 			sb.WriteString(",")
	// 		}
	// 	}
	// 	sb.WriteString(`) `)
	// 	check = true
	// } else if evtLogReq.Code2 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`WHERE event_domain = %d AND event_category = %d `, evtLogReq.Code1, evtLogReq.Code2))
	// 	check = true
	// } else if evtLogReq.Code1 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`WHERE event_domain = %d `, evtLogReq.Code1))
	// 	check = true
	// }

	// if len(evtLogReq.SubTypeArr) != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(`subject_type in(`)
	// 	for idx, sub := range evtLogReq.SubTypeArr {
	// 		sb.WriteString(fmt.Sprintf(`%d`, sub))
	// 		if len(evtLogReq.SubTypeArr)-1 != idx {
	// 			sb.WriteString(",")
	// 		}
	// 	}
	// 	sb.WriteString(`) `)
	// }

	// if evtLogReq.SubId1 != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`subject_id1 = %d `, evtLogReq.SubId1))
	// }

	// if evtLogReq.SubId2 != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`subject_id2 = %d `, evtLogReq.SubId2))
	// }

	// if evtLogReq.TarType != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`target_type = %d `, evtLogReq.TarType))
	// }

	// if evtLogReq.TarId1 != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`target_id1 = %d `, evtLogReq.TarId1))
	// }

	// if evtLogReq.TarId2 != 0 {
	// 	if !check {
	// 		sb.WriteString(`WHERE `)
	// 		check = true
	// 	} else {
	// 		sb.WriteString(`AND `)
	// 	}
	// 	sb.WriteString(fmt.Sprintf(`target_id2 = %d `, evtLogReq.TarId2))
	// }

	// if evtLogReq.Level != -1 {
	// 	var joinEventContent strings.Builder
	// 	joinEventContent.WriteString(fmt.Sprintf(`SELECT b.* FROM event_small a INNER JOIN ( %s ) b
	// 										ON a.event_domain = b.event_domain
	// 										AND a.event_category = b.event_category
	// 										AND a.event_content = b.event_content
	// 										AND a.event_level = %d`, sb.String(), evtLogReq.Level))
	// 	sb = joinEventContent
	// }

	var completeSb strings.Builder
	completeSb.WriteString(fmt.Sprintf(`SELECT c.* FROM ( %s ) c WHERE c.coll_dt BETWEEN ? AND ? ORDER BY coll_dt %s LIMIT ? OFFSET ?`, sb.String(), sortOrder))

	rows, err := i.db.QueryContext(timeout, completeSb.String(), startDT, endDT, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return nil, nil
}

// 중복된 레코드 삽입 시 활용할 수 있는 Native Query
func (i *Maria) InsertTempData(tbName string, objs []any) error {
	// if len(objs) == 0 {
	// 	logger.Infoln("[DB] No temp data to insert")
	// 	return nil
	// }

	// timeout, cancel := context.WithTimeout(i.ctx, i.timeout)
	// defer cancel()

	// dayMonthYear에 따라 테이블 선택
	// var targetTable string
	// if len(temps) > 0 {
	// 	switch temps[0].DayMonthYear.Int16 {
	// 	case 1:
	// 		targetTable = "st_temp_daily"
	// 	case 2:
	// 		targetTable = "st_temp_monthly"
	// 	case 3:
	// 		targetTable = "st_temp_yearly"
	// 	default:
	// 		return fmt.Errorf("invalid dayMonthYear value: %d", temps[0].DayMonthYear.Int16)
	// 	}
	// }

	// var query strings.Builder
	// query.WriteString(fmt.Sprintf(`INSERT INTO %s (station_id, agr_dt, avl_temp, max_temp, min_temp, day_month_year) VALUES `, tbName))

	// placeholders := make([]string, len(objs))
	// args := make([]interface{}, 0, len(objs)*6)

	// for idx, temp := range objs {
	// 	placeholders[idx] = `(?, ?, ?, ?, ?, ?)`
	// 	args = append(args,
	// 		temp.StationId.Int32,
	// 		temp.AgrDt.Time,
	// 		temp.AvlTemp.Float64,
	// 		temp.MaxTemp.Float64,
	// 		temp.MinTemp.Float64,
	// 		temp.DayMonthYear.Int16,
	// 	)
	// }

	// query.WriteString(strings.Join(placeholders, ", "))

	// // 중복 시 업데이트
	// query.WriteString(`
	// 	ON DUPLICATE KEY UPDATE
	// 		avl_temp = VALUES(avl_temp),
	// 		max_temp = VALUES(max_temp),
	// 		min_temp = VALUES(min_temp)
	// `)

	// if _, err := i.db.ExecContext(timeout, query.String(), args...); err != nil {
	// 	logger.Infoln(err)
	// 	return err
	// }

	return nil
}

// 일,월,년 별 집계 Native Query
func (i *Maria) AggregateTempSelect(tableName string, startDt, endDt time.Time, dayMonthYear int) (any, error) {
	timeout, cancel := context.WithTimeout(i.ctx, i.timeout)
	defer cancel()

	var query strings.Builder
	query.WriteString(`
		SELECT b.station_id AS station_id, 
			b.agr_dt AS agr_dt, 
			ROUND(SUM(b.stt4)/COUNT(*), 2) AS avl_temp, 
			MAX(b.stt4) AS max_temp,
			MIN(b.stt4) AS min_temp,
	`)
	query.WriteString(fmt.Sprintf(`%d AS day_month_year `, dayMonthYear))

	// dayMonthYear에 따라 날짜 포맷 변경
	switch dayMonthYear {
	case 1: // 일별: yyyy-mm-dd
		query.WriteString(`FROM (
			SELECT a.*, STR_TO_DATE(DATE_FORMAT(a.coll_dt, '%Y-%m-%d'), '%Y-%m-%d') AS agr_dt FROM `)
	case 2: // 월별: yyyy-mm-01
		query.WriteString(`FROM (
			SELECT a.*, STR_TO_DATE(DATE_FORMAT(a.coll_dt, '%Y-%m-01'), '%Y-%m-%d') AS agr_dt FROM `)
	case 3: // 년별: yyyy-01-01
		query.WriteString(`FROM (
			SELECT a.*, STR_TO_DATE(DATE_FORMAT(a.coll_dt, '%Y-01-01'), '%Y-%m-%d') AS agr_dt FROM `)
	case 4: // 시간별: yyyy-mm-dd hh:00:00
		query.WriteString(`FROM (
			SELECT a.*, STR_TO_DATE(DATE_FORMAT(a.coll_dt, '%Y-%m-%d %H:00:00'), '%Y-%m-%d %H:%i:%s') AS agr_dt FROM `)
	}

	query.WriteString(fmt.Sprintf(`%s a
		) b
		WHERE b.device_list_id = 1
			AND b.coll_dt BETWEEN ? AND ?
			AND b.stt1 = 0 
		GROUP BY b.station_id, b.agr_dt`, tableName))

	rows, err := i.db.QueryContext(timeout, query.String(), startDt, endDt)
	if err != nil {
		logger.Infoln(err)
		return nil, err
	}
	defer rows.Close()

	return nil, nil
}
