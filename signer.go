package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// func main() {

// 	// global vars
// 	// try to use wg
// 	// try to use mu
// }

func ExecutePipeline(freeFlowJobs ...job) {
	in, out := make(chan interface{}), make(chan interface{})
	for index, job := range freeFlowJobs {
		if index%2 == 0 {
			go job(in, out)
		} else {
			go job(out, in)
		}
		time.Sleep(time.Millisecond * 10)
	}
	time.Sleep(time.Millisecond * 2800)
}

func SingleHash(in, out chan interface{}) {
	// fmt.Println("	SIGNLE_HASH")
	var input, crc32, md5, crc32md5, hash string
	var count int
	for data := range in {
		switch data.(type) {
		case int:
			go func(data interface{}, out chan interface{}, count int) {
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
				time.Sleep(time.Millisecond * 1100) /// need to try use less time

				hash = crc32 + "~" + crc32md5

				fmt.Printf("%d SingleHash result: %s\n\n", count, hash)
				out <- hash
			}(data, out, count)
			time.Sleep(time.Millisecond * 12)
			count++
		case string:
			in <- data
			runtime.Gosched()
			// continue
		}
	}

}

// time enough only for 6 interation of MultiHash
func MultiHash(in, out chan interface{}) {

	// fmt.Println("	MULTI_HASH")
	for data := range in {

		fmt.Printf("MultiHash input: %s\n\n", data)
		switch data.(type) {
		case int:
			in <- data
		case string: // do each arr element by goroutine then
			go func(data interface{}, out chan interface{}) {
				var multiHash string
				singleHash := data.(string)
				// go func(singleHash string, multiHash *string) {
				arr := make([]string, 6)
				fmt.Println("TEST TEST TEST")
				// for th := 0; th > 6; th++ {
				// 	fmt.Printf("%d TEST TEST TEST", th)
				go iterMultiHash(singleHash, 0, arr)
				go iterMultiHash(singleHash, 1, arr)
				go iterMultiHash(singleHash, 2, arr)
				go iterMultiHash(singleHash, 3, arr)
				go iterMultiHash(singleHash, 4, arr)
				go iterMultiHash(singleHash, 5, arr)
				// }
				time.Sleep(time.Millisecond * 1050)
				multiHash += strings.Join(arr, "")
				// }(singleHash, &multiHash)
				out <- multiHash
				fmt.Printf("%s MultiHash result: %s\n\n", singleHash, multiHash)
			}(data, out)
		}
	}

}

func iterMultiHash(singleHash string, th int, arr []string) {
	arr[th] = DataSignerCrc32(strconv.Itoa(th) + singleHash) // do it with goroutine
	fmt.Printf("%s MultiHash: crc32(th+step1)): %d %s\n", singleHash, th, arr[th])
}

func CombineResults(in, out chan interface{}) {

	// fmt.Println("	COMBINE_HASH")
	arr := []string{}
	for hash := range in {
		var strHash string // := hash.(string)
		switch hash.(type) {
		case int:
			in <- hash
			runtime.Gosched()
			continue
		case string:
			strHash = hash.(string)
			fmt.Println("c_in:", strHash)
			arr = append(arr, strHash)
		}

	}
	// need to sort

	sort.Strings(arr)
	result := strings.Join(arr, "_")
	// fmt.Println(result)
	out <- result
}

// leftHalf := DataSignerCrc32(data.(string))
// rightHalf := DataSignerCrc32(DataSignerMd5(data.(string)))

/* func concurCall(mu *sync.Mutex, function job, in, out chan interface{}) {
	mu.Lock()
	function(in, out)
	mu.Unlock()
}
*/
