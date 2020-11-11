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

// ExecutePipeline обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
func ExecutePipeline(freeFlowJobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, function := range freeFlowJobs {
		out := make(chan interface{})
		wg.Add(1)

		go work(wg, function, in, out)
		in = out
	}
	wg.Wait()
}

func work(wg *sync.WaitGroup, function job, in, out chan interface{}) {
	defer wg.Done()
	defer close(out)
	function(in, out) // call incoming job
}

// SingleHash читает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~), где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	wgSingleHash := &sync.WaitGroup{} // wait goroutines called in singleHash
	for value := range in {

		// calculate each set of hash in goroutine
		data := fmt.Sprintf("%v", value)
		md5 := DataSignerMd5(data)
		wgSingleHash.Add(1)
		go iterSingleHash(wgSingleHash, out, data, md5)

		time.Sleep(time.Millisecond * 10)
	}
	wgSingleHash.Wait() // Waiting

}

func iterSingleHash(wgSingleHash *sync.WaitGroup, out chan interface{}, data, md5 string) {
	defer wgSingleHash.Done()

	crc32Ch := make(chan string)
	crc32md5Ch := make(chan string)
	go func(crc32Ch chan string, data string) {
		crc32Ch <- DataSignerCrc32(data)
	}(crc32Ch, data)

	go func(crc32md5Ch chan string, data, md5 string) {
		crc32md5Ch <- DataSignerCrc32(md5)
	}(crc32md5Ch, data, md5)

	crc32 := <-crc32Ch
	crc32md5 := <-crc32md5Ch
	singeHash := crc32 + "~" + crc32md5

	out <- singeHash
	// fmt.Printf("%s SingleHash md5(data): %s\n", data, md5)
	// fmt.Printf("%s SingleHash crc32(data): %s\n", data, crc32)
	// fmt.Printf("%s SingleHash crc32(md5(data)): %s\n", data, crc32md5)
	// fmt.Printf("%s SingleHash result: %s\n\n", data, singeHash)
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	wgMultiHash := &sync.WaitGroup{} // wait ending all multiHash goroutines

	for data := range in {
		singleHash := fmt.Sprintf("%v", data)
		// fmt.Printf("MultiHash input: %s\n", singleHash)

		wgMultiHash.Add(1)
		go iterMultiHash(wgMultiHash, out, singleHash)
	}
	wgMultiHash.Wait()
}

func iterMultiHash(wgMultiHash *sync.WaitGroup, out chan interface{}, singleHash string) {
	defer wgMultiHash.Done()
	wgTemp := &sync.WaitGroup{}
	arr := make([]string, 6, 6)

	for th := 0; th < 6; th++ {
		wgTemp.Add(1)
		go func(wgTemp *sync.WaitGroup, arr []string, th int, singleHash string) {
			defer wgTemp.Done()
			arr[th] = DataSignerCrc32(strconv.Itoa(th) + singleHash) // do it with goroutine
			// fmt.Printf("%s MultiHash: crc32(th+step1)): %d %s\n", singleHash, th, arr[th])

		}(wgTemp, arr, th, singleHash)

	}
	wgTemp.Wait()
	multiHash := strings.Join(arr, "")
	out <- multiHash

}

//CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/), объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {

	var sArr []string
	for hash := range in {
		strHash := hash.(string)
		// fmt.Println("c_in:", strHash)
		sArr = append(sArr, strHash)
	}

	sort.Strings(sArr)
	result := strings.Join(sArr, "_")
	// fmt.Println(result)
	out <- result
}
