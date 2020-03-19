package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func openDB() {
	db, err := sql.Open("sqlite3", HomeDir+"/.fb2c/books.db")
	if err != nil {
		panic(err)
	}
	DB = db
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE,
    pos INTEGER
    )`)
	if err != nil {
		panic(err)
	}
}

func createBookDirectory() {
	if _, err := os.Stat(HomeDir + "/.fb2c/"); os.IsNotExist(err) {
		err = os.Mkdir(HomeDir+"/.fb2c/", 0755)
		if err != nil {
			panic(err)
		}
	}
}

type Book struct {
	Id   int
	Path string
	Pos  int
}

var HomeDir string

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	HomeDir = usr.HomeDir
	createBookDirectory()
	openDB()
	if len(os.Args) < 2 {
		fmt.Println("Usage: fb2 <filename>")
		return
	}
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Println("Bad file")
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Bad file")
	}
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	pars := strings.Split(strings.Trim(string(out), "\n"), " ")
	h, _ := strconv.Atoi(pars[0])
	w, _ := strconv.Atoi(pars[1])
	st_ind := 0
	res, err := DB.Query("SELECT * FROM books WHERE path=$1", path)
	if err != nil {

	} else {
		for res.Next() {
			b := Book{}
			err := res.Scan(&b.Id, &b.Path, &b.Pos)
			if err != nil {
				//fmt.Printf("Error during sql query: %v\n", err.Error())
			} else {
				st_ind = b.Pos
			}
		}
	}

	count := h * w / 2 * 2
	data_s := string(data)
	re := regexp.MustCompile(`<[\/]?[\"\'\:\#\=_\w\s\?\-\./]+>`)
	dat := re.ReplaceAllString(data_s, "")
	reader := bufio.NewReader(os.Stdin)
	for {
		DB.Exec("INSERT OR REPLACE INTO books VALUES(NULL, $1, $2)", path, st_ind)
		fmt.Printf(dat[st_ind : st_ind+count])
		//fmt.Println(st_ind)
		char, _, _ := reader.ReadLine()
		s := string(char)
		switch s {
		case "b":
			fmt.Printf("\033[H\033[J")
			st_ind = int(math.Max(float64(st_ind-w/2*2), 0))
			break
		default:
			st_ind = int(math.Min(float64(st_ind+w/2*2), float64(len(data_s))))
		}
	}
}
