package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// why for loop can`t work in goroutine

// ExecutePipeline обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
func ExecutePipeline(freeFlowJobs ...job) {
	in := make(chan interface{})
	// start := time.Now()
	for _, function := range freeFlowJobs {
		out := make(chan interface{})
		// if index == 4 {
		// 	// time.Sleep(time.Millisecond * 2)
		// 	time.Sleep(time.Millisecond * 2000)
		// }
		/* if index%2 == 0 {
			go function(in, out)
		} else {
			go function(out, in)
		} */
		go func(in, out chan interface{}, function job) {
			function(in, out)
			time.Sleep(time.Millisecond * 2500)

			close(out)
		}(in, out, function)

		// fmt.Printf("===========================\n%d job start time: %v\n===========================\n", index, time.Since(start))
		in = out
	}
	time.Sleep(time.Millisecond * 2800)
	// time.Sleep(time.Millisecond * 1500)
}

// SingleHash читает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~), где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	// fmt.Println("	SIGNLE_HASH")
	defer close(out)
	var count int
	for data := range in {
		/* switch data.(type) {
		case int: */
		go func(data interface{}, out chan interface{}, count int) {
			var input, crc32, md5, crc32md5, hash string
			start := time.Now()
			input = strconv.Itoa(data.(int))

			fmt.Printf("%d SingleHash data: %s\n", count, input)

			md5 = DataSignerMd5(input)
			fmt.Printf("%d SingleHash md5(data): %s\n", count, md5)

			go func(crc32 *string, input string, count int) {
				temp := DataSignerCrc32(input)
				*crc32 = temp
				fmt.Printf("%d SingleHash crc32(data): %s\n", count, temp)
			}(&crc32, input, count)
			go func(crc32md5 *string, input, md5 string) {
				temp := DataSignerCrc32(md5)
				fmt.Printf("%d SingleHash crc32(md5(data)): %s\n", count, temp)
				*crc32md5 = temp
			}(&crc32md5, input, md5)
			time.Sleep(time.Millisecond * 1200) /// need to try use less time

			hash = crc32 + "~" + crc32md5 + "_" + strconv.Itoa(count) // test
			// hash = crc32 + "~" + crc32md5

			fmt.Printf("%d SingleHash result: %s exec time: %v\n\n", count, hash, time.Since(start))
			out <- hash
		}(data, out, count)
		time.Sleep(time.Millisecond * 12)
		count++
		/* case string:
			in <- data
			runtime.Gosched()
		} */
	}

}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	defer close(out)
	// fmt.Println("	MULTI_HASH")
	for data := range in {

		/* 	switch data.(type) {
		case int:
			// in <- data
			// runtime.Gosched()

		case string:  */ // do each arr element by goroutine then
		go func(data interface{}, out chan interface{}) {
			fmt.Printf("MultiHash input: %s\n\n", data)
			var multiHash string
			a := strings.Split(data.(string), "_")
			singleHash := a[0]
			arr := make([]string, 6)
			go iterMultiHash(singleHash, 0, arr)
			go iterMultiHash(singleHash, 1, arr)
			go iterMultiHash(singleHash, 2, arr)
			go iterMultiHash(singleHash, 3, arr)
			go iterMultiHash(singleHash, 4, arr)
			go iterMultiHash(singleHash, 5, arr)

			time.Sleep(time.Millisecond * 1050)
			multiHash += strings.Join(arr, "")

			out <- multiHash
			fmt.Printf("%s, #%s MultiHash result: %s\n\n", singleHash, a[1], multiHash)
		}(data, out)
		/* } */
	}

}

func iterMultiHash(singleHash string, th int, arr []string) {
	arr[th] = DataSignerCrc32(strconv.Itoa(th) + singleHash) // do it with goroutine
	fmt.Printf("%s MultiHash: crc32(th+step1)): %d %s\n", singleHash, th, arr[th])
}

//CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/), объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	defer close(out)

	// fmt.Println("	COMBINE_HASH")
	var sArr []string
	// var iArr []string
	for hash := range in {
		var strHash string // := hash.(string)
		/* switch hash.(type) {
		case int:
			in <- hash
			// runtime.Gosched()
			// continue
		case string: */
		strHash = hash.(string)
		fmt.Println("c_in:", strHash)
		sArr = append(sArr, strHash)
		// // iArr = append(iArr, strHash)
		/* } */
		fmt.Printf("\n\nCOMBINE_RESULTS ARR len: %d, value: %v\n\n", len(sArr), sArr)
		// // fmt.Printf("\n\nCOMBINE_RESULTS ARR len: %d, value: %v\n\n", len(iArr), iArr)
	}
	// need to close channels to end range from in
	fmt.Println("TEST TEST TEST")

	sort.Strings(sArr)
	// sort.Strings(iArr)
	fmt.Printf("\n\nCOMBINE_RESULTS SORTED ARR value: %v\n\n", sArr)
	result := strings.Join(sArr, "_")
	// result := strings.Join(iArr, "_")
	// fmt.Println(result)
	out <- result
	// in <- result
}
