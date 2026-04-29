package oracle

import (
	"context"
	"fmt"
	"pjt/internal/logger"
	"strings"
	"time"
)

// 확장 INSERT문 사용
func (i *Oracle) InsertEventLog(tableName string, params [][]any) error {
	timeout, cancel := context.WithTimeout(i.ctx, i.timeout)
	defer cancel()

	var query strings.Builder
	query.WriteString(`INSERT ALL `)
	for _, param := range params {
		nParam := append([]any{tableName}, param...)
		query.WriteString(fmt.Sprintf(`INTO %s (param...) 
		 VALUES (%d, %d, %d, to_date('%s','YYYY-MM-DD HH24:MI:SS'),%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f) `, nParam...))
	}
	query.WriteString("SELECT * FROM dual")

	if _, err := i.db.ExecContext(timeout, query.String()); err != nil {
		logger.Infoln(err)
		return err
	}

	return nil
}

// Native Query, Paging
func (o *Oracle) SelectEvtLogPaged(tableName string, startDT, endDT time.Time, evtLogReq map[any]any) (any, error) {
	timeout, cancel := context.WithTimeout(o.ctx, o.timeout)
	defer cancel()

	page := 1
	size := 100
	order := 1

	startRow := (page-1)*size + 1
	// endRow에 1을 더하는 이유는 다음 페이지가 존재하는지 확인하기 위함.
	// e.g. pageSize: 100, page: 2 -> startRow: 101, endRow: 201
	endRow := page*size + 1

	var sortOrder string
	if order == 0 {
		sortOrder = `ASC`
	} else {
		sortOrder = `DESC`
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`
            									SELECT *
            									FROM %s t
            									WHERE `, tableName))

	// 아래는 조건 검색 및 조인 -> 상황에 맞게 사용
	// if len(evtLogReq.CodeArr3) != 0 {
	// 	sb.WriteString(fmt.Sprintf(`t.event_domain = %d AND t.event_category = %d AND t.event_content in( `, evtLogReq.Code1, evtLogReq.Code2))
	// 	for idx, code3 := range evtLogReq.CodeArr3 {
	// 		sb.WriteString(fmt.Sprintf(`%d`, code3))
	// 		if len(evtLogReq.CodeArr3)-1 != idx {
	// 			sb.WriteString(",")
	// 		}
	// 	}
	// 	sb.WriteString(`) AND `)
	// } else if evtLogReq.Code2 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`t.event_domain = %d AND t.event_category = %d AND `, evtLogReq.Code1, evtLogReq.Code2))
	// } else if evtLogReq.Code1 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`t.event_domain = %d AND `, evtLogReq.Code1))
	// }

	// sb.WriteString(`t.coll_dt BETWEEN :1 AND :2  `)

	// if evtLogReq.TarType != 0 {
	// 	sb.WriteString(fmt.Sprintf(`AND t.target_type = %d `, evtLogReq.TarType))
	// }
	// if evtLogReq.TarId1 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`AND t.target_id1 = %d `, evtLogReq.TarId1))
	// }
	// if evtLogReq.TarId2 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`AND t.target_id2 = %d `, evtLogReq.TarId2))
	// }
	// if len(evtLogReq.SubTypeArr) != 0 {
	// 	sb.WriteString(`AND t.subject_type in(`)
	// 	for idx, sub := range evtLogReq.SubTypeArr {
	// 		sb.WriteString(fmt.Sprintf(`%d`, sub))
	// 		if len(evtLogReq.SubTypeArr)-1 != idx {
	// 			sb.WriteString(",")
	// 		}
	// 	}
	// 	sb.WriteString(`) `)
	// }
	// if evtLogReq.SubId1 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`AND t.target_id1 = %d `, evtLogReq.SubId1))
	// }
	// if evtLogReq.SubId2 != 0 {
	// 	sb.WriteString(fmt.Sprintf(`AND t.target_id2 = %d `, evtLogReq.SubId2))
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
	completeSb.WriteString(
		fmt.Sprintf(`SELECT * FROM (SELECT a.*, ROW_NUMBER() OVER (ORDER BY a.coll_dt %s) AS rn FROM ( %s ) a ) WHERE rn BETWEEN :3 AND :4`,
			sortOrder, sb.String()))
	rows, err := o.db.QueryContext(timeout, completeSb.String(), startDT, endDT, startRow, endRow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return nil, nil
}

// 중복된 레코드 삽입 시 활용할 수 있는 Native Query
func (i *Oracle) InsertTempData(tbName string, objs []any) error {

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

	// for _, temp := range objs {
	// 	query := fmt.Sprintf(`
	// 		BEGIN
	// 			INSERT INTO %s (station_id, agr_dt, avl_temp, max_temp, min_temp, day_month_year)
	// 			VALUES (%d, TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS'), %.2f, %.2f, %.2f, %d);
	// 		EXCEPTION
	// 			WHEN DUP_VAL_ON_INDEX THEN
	// 				UPDATE %s SET
	// 					avl_temp = %.2f,
	// 					max_temp = %.2f,
	// 					min_temp = %.2f
	// 				WHERE station_id = %d
	// 					AND agr_dt = TO_DATE('%s', 'YYYY-MM-DD HH24:MI:SS')
	// 					AND day_month_year = %d;
	// 		END;
	// 	`,
	// 		tbName,
	// 		temp.StationId.Int32,
	// 		temp.AgrDt.Time.Format(time.DateTime),
	// 		temp.AvlTemp.Float64,
	// 		temp.MaxTemp.Float64,
	// 		temp.MinTemp.Float64,
	// 		temp.DayMonthYear.Int16,
	// 		tbName,
	// 		temp.AvlTemp.Float64,
	// 		temp.MaxTemp.Float64,
	// 		temp.MinTemp.Float64,
	// 		temp.StationId.Int32,
	// 		temp.AgrDt.Time.Format(time.DateTime),
	// 		temp.DayMonthYear.Int16)

	// 	if _, err := i.db.ExecContext(timeout, query); err != nil {
	// 		logger.Infoln(err)
	// 		return err
	// 	}
	// }

	return nil
}

// 일,월,년 별 집계 Native Query
func (i *Oracle) AggregateTempSelect(tableName string, startDt, endDt time.Time, dayMonthYear int) (any, error) {
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
			SELECT a.*, TO_DATE(TO_CHAR(a.coll_dt, 'YYYY-MM-DD'), 'YYYY-MM-DD') AS agr_dt FROM `)
	case 2: // 월별: yyyy-mm-01
		query.WriteString(`FROM (
			SELECT a.*, TO_DATE(TO_CHAR(a.coll_dt, 'YYYY-MM') || '-01', 'YYYY-MM-DD') AS agr_dt FROM `)
	case 3: // 년별: yyyy-01-01
		query.WriteString(`FROM (
			SELECT a.*, TO_DATE(TO_CHAR(a.coll_dt, 'YYYY') || '-01-01', 'YYYY-MM-DD') AS agr_dt FROM `)
	case 4: // 시간별: yyyy-mm-dd hh24:00:00
		query.WriteString(`FROM (
			SELECT a.*, TO_DATE(TO_CHAR(a.coll_dt, 'YYYY-MM-DD HH24') || ':00:00', 'YYYY-MM-DD HH24:MI:SS') AS agr_dt FROM `)
	}

	query.WriteString(fmt.Sprintf(`%s a
		) b
		WHERE b.device_list_id = 1
			AND b.coll_dt BETWEEN :1 AND :2
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
