package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type myCloser interface {
	Close() error
}

type Bills struct {
	TotalAmount float64
	TotalItems int
	AI map[string][]string //AmountAndItem: "filename": [amount, items]
}

var b Bills

func (bs *Bills) cal() {
	as, is := 0.0, 0

	for _, item := range bs.AI {
		a, _ := strconv.ParseFloat(item[0], 64)
		i, _ := strconv.Atoi(item[1])
		as += a
		is += i
	}
	bs.TotalAmount = as
	bs.TotalItems = is
}

func (bs Bills) String() string {
	bs.cal()
	s := fmt.Sprintf("总金额：%f 订单数：%d\n", bs.TotalAmount, bs.TotalItems)
	for filename, item := range bs.AI {
		s += fmt.Sprintf("%s, %s, %s\n", filename, item[0], item[1])
		// a, _ := strconv.ParseFloat(item[0], 64)
		// i, _ := strconv.Atoi(item[1])
		// as += a
		// is += i
	}
	return s
}

// closeFile is a helper function which streamlines closing
// with error checking on different file types.
func closeFile(f myCloser) {
	err := f.Close()
	check(err)
}

// readAll is a wrapper function for ioutil.ReadAll. It accepts a zip.File as
// its parameter, opens it, reads its content and returns it as a byte slice.
func readAll(file *zip.File) []byte {
	fc, err := file.Open()
	check(err)
	defer closeFile(fc)

	content, err := ioutil.ReadAll(fc)
	check(err)

	return content
}

// write slick to file
func write(slice []string, out string) {
	ioutil.WriteFile(out, []byte(strings.Join(slice, "\n")), 0644)
}

// check is a helper function which streamlines error checking
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func use(v interface{}) {}

func readOrders(file *zip.File) (orders []string) {
	fc, err := file.Open()
	check(err)
	defer closeFile(fc)

	scanner := bufio.NewScanner(fc)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		orders = append(orders, fields[11])
	}
	return
}

func readAI(file *zip.File) (amount string, items string) {
	fc, err := file.Open()
	check(err)
	defer closeFile(fc)

	scanner := bufio.NewScanner(fc)
	scanner.Split(bufio.ScanLines)
	//nu := 0
	for scanner.Scan() {
		
		line := scanner.Text()
		//转码
		i:= bytes.NewReader([]byte(line))
		decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
		bts, _ := ioutil.ReadAll(decoder)
		line = string(bts)
		//fmt.Println(line)

		if strings.HasPrefix(line, "总计") {
			fields := strings.Fields(line)
			if fields[2] != "0.00" {
				fmt.Printf("%s 借：%s\n", file.Name, fields[2])
			}
			return fields[3], fields[1]
		}
		// nu++
		// line := scanner.Text()
		// if nu == 14 {
		// 	fields := strings.Fields(line)
		// 	if fields[2] != "0.00" {
		// 		fmt.Printf("%s 借：%s\n", file.Name, fields[2])
		// 	}
		// 	return fields[3], fields[1]
		// }
	}
	return "0.0", "0"
}

// read bills from zip file
func readBillsFromZip(zipFile string) (orders []string, amount string, items string) {
	zf, err := zip.OpenReader(zipFile)
	check(err)
	defer closeFile(zf)

	for _, file := range zf.File {
		// 明细
		if strings.HasPrefix(file.Name, "INN") {
			orders = readOrders(file)
		}
		// 汇总
		if strings.HasPrefix(file.Name, "RD") {
			//TODO
			amount, items = readAI(file)
		}

	}
	return
}
//从文件夹导出流水号(包含对账zip文件)
func exportFolder(dir string, out string) {
	fis, err := ioutil.ReadDir(dir)
	check(err)
	orders := make([]string, 0)
	for _, fi := range fis {
		zipFile := dir + "/" + fi.Name()
		o, a, i := readBillsFromZip(zipFile)
		//fmt.Println(fi.Name(), a, i, "=", len(o), o)
		b.AI[fi.Name()] = []string{a, i}
		//orders = append(orders, fi.Name())
		orders = append(orders, o...)
	}
	write(orders, out)
}

func exportSingle(zipFile string, out string) {
	o, a, i := readBillsFromZip(zipFile)
	//fmt.Println(zipFile, len(o), o)
	b.AI[zipFile] = []string{a, i}
	write(o, out)
}

func init() {
	b = Bills{AI: make(map[string][]string)}
}
func main() {
	exportFolder("./zip_p", "jc0706.txt")
	fmt.Println(b)
	// fmt.Println(strings.HasPrefix("总计 fdsf", "总计"))
	// fmt.Println(strings.HasPrefix("fdsf", "fd"))
}
