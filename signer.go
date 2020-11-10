package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// why for loop can`t work in goroutine
var mu = &sync.Mutex{}

// ExecutePipeline обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
func ExecutePipeline(freeFlowJobs ...job) {
	in := make(chan interface{})
	for _, function := range freeFlowJobs {
		out := make(chan interface{})

		go func(in, out chan interface{}, function job, mu *sync.Mutex) {
			// mu.Lock()
			function(in, out)
			// mu.Unlock()
			time.Sleep(time.Millisecond * 2500)
			close(out)
		}(in, out, function, mu)

		in = out
	}
	time.Sleep(time.Millisecond * 2800)
}

// SingleHash читает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~), где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	defer close(out)
	var count int
	for data := range in {

		go func(data interface{}, out chan interface{}, count int, mu *sync.Mutex) {
			var input, crc32, md5, crc32md5, hash string
			// start := time.Now()
			input = strconv.Itoa(data.(int))
			// fmt.Printf("%d SingleHash data: %s\n", count, input)

			mu.Lock()
			md5 = DataSignerMd5(input)
			mu.Unlock()
			// fmt.Printf("%d SingleHash md5(data): %s\n", count, md5)

			go func(crc32 *string, input string, count int, mu *sync.Mutex) {
				mu.Lock()
				temp := DataSignerCrc32(input)
				*crc32 = temp
				mu.Unlock()
				// fmt.Printf("%d SingleHash crc32(data): %s\n", count, temp)
			}(&crc32, input, count, mu)
			go func(crc32md5 *string, input, md5 string, mu *sync.Mutex) {
				mu.Lock()
				temp := DataSignerCrc32(md5)
				// fmt.Printf("%d SingleHash crc32(md5(data)): %s\n", count, temp)
				*crc32md5 = temp
				mu.Unlock()
			}(&crc32md5, input, md5, mu)
			time.Sleep(time.Millisecond * 1200) /// need to try use less time

			// hash = crc32 + "~" + crc32md5 + "_" + strconv.Itoa(count) // test
			mu.Lock()
			hash = crc32 + "~" + crc32md5
			mu.Unlock()

			// fmt.Printf("%d SingleHash result: %s exec time: %v\n\n", count, hash, time.Since(start))
			// last changes
			mu.Lock()
			out <- hash
			mu.Unlock()
			// last changes

		}(data, out, count, mu)
		time.Sleep(time.Millisecond * 12)
		count++
	}

}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	defer close(out)
	for data := range in {

		go func(data interface{}, out chan interface{}, mu *sync.Mutex) {
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
			mu.Lock()
			multiHash += strings.Join(arr, "")
			mu.Unlock()

			out <- multiHash
			// fmt.Printf("%s, #%s MultiHash result: %s\n\n", singleHash, a[1], multiHash)
		}(data, out, mu)
	}

}

func iterMultiHash(singleHash string, th int, arr []string /* , mu *sync.Mutex */) {
	// mu.Lock()
	arr[th] = DataSignerCrc32(strconv.Itoa(th) + singleHash) // do it with goroutine
	// mu.Unlock()
	// fmt.Printf("%s MultiHash: crc32(th+step1)): %d %s\n", singleHash, th, arr[th])
}

//CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/), объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	defer close(out)

	// fmt.Println("	COMBINE_HASH")
	var sArr []string
	// var iArr []string
	for hash := range in {
		strHash := hash.(string)
		fmt.Println("c_in:", strHash)
		sArr = append(sArr, strHash)

		// fmt.Printf("\n\nCOMBINE_RESULTS ARR len: %d, value: %v\n\n", len(sArr), sArr)
	}
	// need to close channels to end range from in
	// fmt.Println("TEST TEST TEST")

	sort.Strings(sArr)
	fmt.Printf("\n\nCOMBINE_RESULTS SORTED ARR value: %v\n\n", sArr)
	result := strings.Join(sArr, "_")
	// fmt.Println(result)
	out <- result
}
