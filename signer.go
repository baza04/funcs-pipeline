package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {

	// global vars
	// try to use wg
	// try to use mu
}

func ExecutePipeline(freeFlowJobs ...job) {
	in, out := make(chan interface{}), make(chan interface{})
	for index, job := range freeFlowJobs {
		if index%2 == 0 {
			go job(in, out)
		} else {
			go job(out, in)
		}
		// time.Sleep(time.Millisecond * 10)
	}
	time.Sleep(time.Second * 4)
}

func SingleHash(in, out chan interface{}) {
	// fmt.Println("	SIGNLE_HASH")
	var input, crc32, md5, crc32md5, hash string
	var count int
	for data := range in {
		switch data.(type) {
		case int:
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
			time.Sleep(time.Millisecond * 1050)

			hash = crc32 + "~" + crc32md5

			fmt.Printf("%d SingleHash result: %s\n\n", count, hash)
			out <- hash
			count++
		case string:
			in <- data
			runtime.Gosched()
			// continue
		}
	}
}

func MultiHash(in, out chan interface{}) {

	// fmt.Println("	MULTI_HASH")
	var count int
	for data := range in {
		var multiHash string
		fmt.Printf("MultiHash input: %s\n\n", data)
		switch data.(type) {
		case int:
			in <- data
			continue
		case string:
			singleHash := data.(string)
			go func(singleHash string, multiHash *string) {
				for th := 0; th > 6; th++ {
					temp := DataSignerCrc32(strconv.Itoa(th) + singleHash)
					fmt.Printf("%s MultiHash: crc32(th+step1)): %d %s\n", singleHash, th, temp)
					*multiHash += temp
				}
			}(singleHash, &multiHash)
			out <- multiHash
			fmt.Printf("%s MultiHash result: %s\n\n", singleHash, multiHash)
			count++
		}
	}

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
		}
		fmt.Println("c_in:", strHash)
		arr = append(arr, strHash)

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
