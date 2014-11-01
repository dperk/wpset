package main

import (
	"database/sql"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/coopernurse/gorp"

	_ "github.com/mattn/go-sqlite3"
)

const dbpath = "/Library/Application Support/Dock/desktoppicture.db"

// const dbpath = "/Downloads/desktoppicture.db"

func main() {
	wp := os.Args[1]
	img, err := filepath.Abs(wp)

	checkErr(err, "Not enough args")

	if !isValidImage(img) {
		panic("Not a valid image: " + img)
	}

	dbmap := initDb()
	defer dbmap.Db.Close()
	//
	var wallpapers []Data
	_, err = dbmap.Select(&wallpapers, "SELECT value, ROWID FROM data ORDER BY ROWID ASC")
	checkErr(err, "Select failed")
	log.Println("All Rows:")
	for _, p := range wallpapers[1:] {
		// log.Printf("    %d: %v\n", x, p)

		if isValidImage(expandPath(p.Image)) {
			p.Image = collapsePath(img)
			_, err = dbmap.Update(&p)
			checkErr(err, "Update failed")
			// log.Println("Rows updated: ", cnt)
		}
	}

	cmd := exec.Command("killall", "Dock")
	_, err = cmd.Output()
	checkErr(err, "Error killing dock")
}

type Data struct {
	Id    int32  `db:"ROWID"`
	Image string `db:"value"`
}

func initDb() *gorp.DbMap {
	url := "file:" + homedir() + dbpath + "?cache=shared&mode=rwc"
	db, err := sql.Open("sqlite3", url)
	checkErr(err, "sql.Open failed")

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	// dbmap.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
	dbmap.AddTableWithName(Data{}, "data").SetKeys(false, "Id")

	return dbmap
}

func expandPath(path string) string {
	p := strings.Replace(path, "~", homedir(), 1)
	return p
}

func collapsePath(path string) string {
	p := strings.Replace(path, homedir(), "~", 1)
	return p
}

func homedir() string {
	usr, _ := user.Current()
	home := usr.HomeDir
	return home
}

func isValidImage(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	ext := filepath.Ext(path)
	switch {
	case strings.EqualFold(ext, ".jpg"):
		return true
	}
	return false
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
