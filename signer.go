package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

func main() {

	// global vars
	// try to use wg
	// try to use mu
}

func ExecutePipeline(freeFlowJobs ...job) {
	in, out := make(chan interface{}, 5), make(chan interface{}, 5)
	for index, function := range freeFlowJobs {
		if index%2 == 0 {
			go function(in, out)
		} else {
			go function(out, in)
		}
		// if index == 1 {
		// 	time.Sleep(time.Second * 1)
		// }
		// time.Sleep(time.Millisecond * 250)
	}
	time.Sleep(time.Second * 2)
}

func SingleHash(in, out chan interface{}) {
	// fmt.Println("	SIGNLE_HASH:", <-or)

	fmt.Println("	SIGNLE_HASH")
	var input, crc32, md5, crc32md5, hash string
	for data := range in {
		input = strconv.Itoa(data.(int))
		fmt.Println("s_in:", input)

		md5 = DataSignerMd5(input)
		fmt.Println("md5:", md5)

		go func(crc32 *string, input string) {
			*crc32 = DataSignerCrc32(input)
		}(&crc32, input)
		go func(crc32md5 *string, input, md5 string) {
			fmt.Println("md5 in GO:", md5)
			*crc32md5 = DataSignerCrc32(md5)
		}(&crc32md5, input, md5)
		time.Sleep(time.Millisecond * 1200)
		// if
		hash = crc32 + "~" + crc32md5

		fmt.Printf("crc32_go: %s\ncrc32md5: %s\nhash: %s\n\n", crc32, crc32md5, hash)
		out <- hash
	}
}

func MultiHash(in, out chan interface{}) {

	fmt.Println("	MULTI_HASH")
	th := 0
	var multiHash string
	for data := range in {
		fmt.Println("m_in:", data)
		go func(data interface{}, multiHash *string, th int) {
			*multiHash = DataSignerCrc32(strconv.Itoa(th) + data.(string))
		}(data, &multiHash, th)
		// multiHash := DataSignerCrc32(strconv.Itoa(th) + strconv.Itoa(data.(int)))
		fmt.Println("MyMultiHash:", multiHash)
		out <- multiHash
		th++
	}

}

func CombineResults(in, out chan interface{}) {

	fmt.Println("	COMBINE_HASH")
	arr := []string{}
	for hash := range in {
		var strHash string // := hash.(string)
		switch hash.(type) {
		case int:
			strHash = strconv.Itoa(hash.(int))
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
