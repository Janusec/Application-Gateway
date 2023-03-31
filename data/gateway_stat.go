/*
 * @Copyright Reserved By Janusec (https://www.janusec.com/).
 * @Author: U2
 * @Date: 2020-10-07 09:43:07
 * @Last Modified: U2, 2020-10-07 09:43:07
 */

package data

import (
	"janusec/models"
	"janusec/utils"
)

// CreateTableIfNotExistsAccessStats create statistics table
func (dal *MyDAL) CreateTableIfNotExistsAccessStats() error {
	const sqlCreateTableIfNotExistsStats = `CREATE TABLE IF NOT EXISTS "access_stats"("id" bigserial PRIMARY KEY, "app_id" bigint, "url_path" VARCHAR(256) NOT NULL, "stat_date" VARCHAR(16) NOT NULL, "amount" bigint, "update_time" bigint,  CONSTRAINT "stat_id" unique ("app_id","url_path","stat_date"))`
	_, err := dal.db.Exec(sqlCreateTableIfNotExistsStats)
	return err
}

// IncAmount update access statistics
func (dal *MyDAL) IncAmount(appID int64, urlPath string, statDate string, delta int64, updateTime int64) error {
	//var id, amount int64
	if len(urlPath) > 255 {
		urlPath = urlPath[0:255]
	}
	snakeID := utils.GenSnowflakeID()
	const sql = `INSERT INTO "access_stats"("id","app_id","url_path","stat_date","amount","update_time") VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT ("app_id","url_path","stat_date") DO UPDATE SET "amount"="access_stats"."amount"+$5,"update_time"=$6`
	_, err := dal.db.Exec(sql, snakeID, appID, urlPath, statDate, delta, updateTime)
	if err != nil {
		utils.DebugPrintln("IncAmount insert", err)
	}
	/*
		const sql = `select "id","amount" from "access_stats" where "app_id"=$1 and "url_path"=$2 and "stat_date"=$3 LIMIT 1`
		err := dal.db.QueryRow(sql, appID, urlPath, statDate).Scan(&id, &amount)
		if err != nil {
			// Not existed before
			const sqlInsert = `INSERT INTO "access_stats"("app_id","url_path","stat_date","amount","update_time") VALUES($1,$2,$3,$4,$5)`
			_, err = dal.db.Exec(sqlInsert, appID, urlPath, statDate, delta, updateTime)
			if err != nil {
				utils.DebugPrintln("IncAmount insert", err)
			}
			return err
		}
		const sqlUpdate = `UPDATE "access_stats" SET "amount"=$1,"update_time"=$2 WHERE "id"=$3`
		_, err = dal.db.Exec(sqlUpdate, amount+delta, updateTime, id)
		if err != nil {
			utils.DebugPrintln("IncAmount update", err)
		}
	*/
	return err
}

// ClearExpiredAccessStats clear access statistics before designated time
func (dal *MyDAL) ClearExpiredAccessStats(expiredTime int64) error {
	const sqlDel = `DELETE FROM "access_stats" WHERE "update_time"<$1`
	_, err := dal.db.Exec(sqlDel, expiredTime)
	if err != nil {
		utils.DebugPrintln("ClearExpiredAccessStats", err)
	}
	return err
}

// GetAccessStatByAppIDAndDate return the amount of designated app
func (dal *MyDAL) GetAccessStatByAppIDAndDate(appID int64, statDate string) int64 {
	amount := int64(0)
	if appID == 0 {
		const sqlQuery0 = `SELECT SUM("amount") from "access_stats" WHERE "stat_date"=$1`
		_ = dal.db.QueryRow(sqlQuery0, statDate).Scan(&amount)
		return amount
	}
	const sqlQuery1 = `SELECT SUM("amount") from "access_stats" WHERE "app_id"=$1 and "stat_date"=$2`
	_ = dal.db.QueryRow(sqlQuery1, appID, statDate).Scan(&amount)
	return amount
}

// GetPopularContent return top visited URL Path
func (dal *MyDAL) GetPopularContent(appID int64, statDate string) ([]*models.PopularContent, error) {
	topPaths := []*models.PopularContent{}
	if appID == 0 {
		const sqlQuery0 = `SELECT "app_id","url_path","amount" FROM "access_stats" WHERE "stat_date"=$1 ORDER BY "amount" DESC LIMIT 100`
		rows, _ := dal.db.Query(sqlQuery0, statDate)
		for rows.Next() {
			var popContent = &models.PopularContent{}
			_ = rows.Scan(&popContent.AppID, &popContent.URLPath, &popContent.Amount)
			topPaths = append(topPaths, popContent)
		}
		return topPaths, nil
	}
	const sqlQuery1 = `SELECT "app_id","url_path","amount" from "access_stats" WHERE "app_id"=$1 and "stat_date"=$2 ORDER BY "amount" DESC LIMIT 100`
	rows, _ := dal.db.Query(sqlQuery1, appID, statDate)
	for rows.Next() {
		var popContent = &models.PopularContent{}
		_ = rows.Scan(&popContent.AppID, &popContent.URLPath, &popContent.Amount)
		topPaths = append(topPaths, popContent)
	}
	return topPaths, nil
}

// The following is for Referer Statistics

// CreateTableIfNotExistsRefererStats ...
func (dal *MyDAL) CreateTableIfNotExistsRefererStats() error {
	const sqlCreateTableIfNotExistsStats = `CREATE TABLE IF NOT EXISTS "referer_stats"("id" bigserial PRIMARY KEY, "app_id" bigint, "host" VARCHAR(256) NOT NULL, "url" VARCHAR(256) NOT NULL, "client_id" VARCHAR(128) NOT NULL, "count" bigint, "date_timestamp" bigint, CONSTRAINT "refer_id" unique("app_id", "host", "url", "client_id", "date_timestamp"))`
	_, err := dal.db.Exec(sqlCreateTableIfNotExistsStats)
	return err
}

// UpdateRefererStat ...
func (dal *MyDAL) UpdateRefererStat(appID int64, host string, path string, clientID string, deltaCount int64, dateTimestamp int64) error {
	if len(host) > 250 {
		host = host[:250]
	}
	if len(path) > 250 {
		path = path[:250]
	}
	const sqlUpdateReferStat = `INSERT INTO "referer_stats"("app_id", "host", "url", "client_id", "count", "date_timestamp") VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT ("app_id", "host", "url", "client_id", "date_timestamp") DO UPDATE SET "count"="referer_stats"."count"+$5`
	_, err := dal.db.Exec(sqlUpdateReferStat, appID, host, path, clientID, deltaCount, dateTimestamp)
	return err
}

// ClearExpiredReferStat clear expired stats
func (dal *MyDAL) ClearExpiredReferStat(expiredTime int64) error {
	const sqlClearRefererStat = `DELETE FROM "referer_stats" WHERE "date_timestamp"<$1`
	_, err := dal.db.Exec(sqlClearRefererStat, expiredTime)
	if err != nil {
		utils.DebugPrintln("ClearExpiredReferStat", err)
	}
	return err
}

// GetRefererHosts ...
func (dal *MyDAL) GetRefererHosts(appID int64, statTime int64) (topReferers []*models.RefererHost, err error) {
	if appID == 0 {
		const sqlStatAll = `SELECT "host",SUM("count") AS "total_pv",COUNT(DISTINCT "client_id") FROM "referer_stats" WHERE "date_timestamp">$1 GROUP BY "host" ORDER BY "total_pv" DESC LIMIT 100`
		rows, _ := dal.db.Query(sqlStatAll, statTime)
		for rows.Next() {
			var RefererHost = &models.RefererHost{}
			_ = rows.Scan(&RefererHost.Host, &RefererHost.PV, &RefererHost.UV)
			topReferers = append(topReferers, RefererHost)
		}
		return topReferers, nil
	}
	// appID not 0
	const sqlStatByAPPID = `SELECT "host",SUM("count") AS "total_pv",COUNT(DISTINCT "client_id") FROM "referer_stats" WHERE "app_id"=$1 AND "date_timestamp">$2 GROUP BY "host" ORDER BY "total_pv" DESC LIMIT 100`
	rows, _ := dal.db.Query(sqlStatByAPPID, appID, statTime)
	for rows.Next() {
		var RefererHost = &models.RefererHost{}
		_ = rows.Scan(&RefererHost.Host, &RefererHost.PV, &RefererHost.UV)
		topReferers = append(topReferers, RefererHost)
	}
	return topReferers, nil
}

// GetRefererURLs ...
func (dal *MyDAL) GetRefererURLs(appID int64, host string, statTime int64) (topRefererURLs []*models.RefererURL, err error) {
	if appID == 0 {
		const sqlStatAll = `SELECT "url",SUM("count") AS "total_pv",COUNT(DISTINCT "client_id") FROM "referer_stats" WHERE "host"=$1 and "date_timestamp">$2 GROUP BY "url" ORDER BY "total_pv" DESC LIMIT 100`
		rows, _ := dal.db.Query(sqlStatAll, host, statTime)
		for rows.Next() {
			var RefererURL = &models.RefererURL{}
			_ = rows.Scan(&RefererURL.URL, &RefererURL.PV, &RefererURL.UV)
			topRefererURLs = append(topRefererURLs, RefererURL)
		}
		return topRefererURLs, nil
	}
	// appID not 0
	const sqlStatByAPPID = `SELECT "url",SUM("count") AS "total_pv",COUNT(DISTINCT "client_id") FROM "referer_stats" WHERE "app_id"=$1 AND "host"=$2 AND "date_timestamp">$3 GROUP BY "url" ORDER BY "total_pv" DESC LIMIT 100`
	rows, _ := dal.db.Query(sqlStatByAPPID, appID, host, statTime)
	for rows.Next() {
		var RefererURL = &models.RefererURL{}
		_ = rows.Scan(&RefererURL.URL, &RefererURL.PV, &RefererURL.UV)
		topRefererURLs = append(topRefererURLs, RefererURL)
	}
	return topRefererURLs, nil
}
