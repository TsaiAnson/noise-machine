package main

import (
    "io"
    "io/ioutil"
    "log"
    "math"
    "net/http"
    "os"
    "runtime"
    "strconv"
    "sync"
    "time"
)

func round(x float64) int {
    t := math.Trunc(x)
    if math.Abs(x-t) >= 0.5 {
        return int(t + math.Copysign(1, x))
    }
    return int(t)
}

func cpu(rate float64) {
    done := make(chan int)
    // One for each core
    for i := 0; i < runtime.NumCPU(); i++ {
        go func() {
            for {
                select {
                case <-done:
                    return
                default:
                    // Sleep (rate control)
                    time.Sleep(time.Nanosecond * time.Duration(round(1000000000/rate)))
                }
            }
        }()
    }
}

func mem (instances int) {
    // To hog memory, we will read X instances of a 10MB file and occasionally poke it to ensure that it stays in RAM

    // Downloading file
    f, err := os.Create("memoryInstance")
    if err != nil {
        log.Fatal(err)
    }
    resp, err := http.Get("http://ipv4.download.thinkbroadband.com/10MB.zip")
    if err != nil {
        log.Fatal(err)
    }
    _, err = io.Copy(f, resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    resp.Body.Close()
    f.Close()

    // Reading instantiations into memory
    files := make([][]byte, instances)
    for i := 0; i < instances; i++ {
        files[i], err = ioutil.ReadFile("memoryInstance")
        if err != nil {
            log.Fatal(err)
        }
    }

    // Poke files every second
    for {
        time.Sleep(time.Second)

        for i := 0; i < instances; i++ {
            test := files[i]
            files[i] = test
        }
    }
}

func disk(rate float64) {
    for {
        // Create + write file
        f, err := os.Create("test.txt")
        if err != nil {
            log.Fatal(err)
        }
        if err := f.Truncate(1e6); err != nil {
            log.Fatal(err)
        }
        f.Close()

        // Read + write file
        data, err := ioutil.ReadFile("test.txt")
        if err != nil {
            log.Fatal(err)
        }
        f, err = os.Open("test.txt")
        if err != nil {
            log.Fatal(err)
        }
        f.Write(data)
        f.Close()

        // Remove file
        err = os.Remove("test.txt")
        if err != nil {
            log.Fatal(err)

        }

        // Sleep (rate control)
        time.Sleep(time.Nanosecond * time.Duration(round(1000000000/rate)))
    }
}

func net(conc int, rate float64) {
    for i := 0; i < conc; i++ {
        go func() {
            for {
                _,_ = http.Get("http://noiseserver.q:80/static/10MB.zip")
                // Sleep (rate control)
                time.Sleep(time.Nanosecond * time.Duration(round(1000000000/rate)))
            }
        }()
    }

    // Kept for debugging
    //for {
    //    resp, err := http.Get("http://noiseserver.q:80/static/512MB.zip")
    //    if resp != nil && err == nil {
    //        // Downloading file
    //        f, err := os.Create("memoryInstance")
    //        if err != nil {
    //            log.Fatal(err)
    //        }
    //        _, err = io.Copy(f, resp.Body)
    //        if err != nil {
    //            log.Fatal(err)
    //        }
    //        resp.Body.Close()
    //        f.Close()
    //    } else {
    //        fmt.Println("Initializing...")
    //    }
    //    time.Sleep(2 * time.Minute)
    //}
}

func main() {
    // Parsing environment variables
    // Should be in terms of per second (range: float64 >0 to 1000000000, any higher will just use whole CPU)
    cpuRate, err := strconv.ParseFloat(os.Getenv("CPU"), 64)
    if err != nil {
        log.Fatal("Unable to parse CPU rate.")
    }

    // Should be in terms of 10MB (ie 5 is equivalent to 50MB of memory)
    memInstance, err := strconv.Atoi(os.Getenv("MEM"))
    if err != nil {
        log.Fatal("Unable to parse MEM rate.")
    }

    // Should be in terms of actions per second (range: float64 >0 to 1000000000)
    diskRate, err := strconv.ParseFloat(os.Getenv("DISK"), 64)
    if err != nil {
        log.Fatal("Unable to parse DISK rate.")
    }

    // Should be number of concurrent downloads
    netConc, err := strconv.Atoi(os.Getenv("NETCONC"))
    if err != nil {
        log.Fatal("Unable to parse NET conc.")
    }

    // Should be downloads per minute (range: float64 >0 to 1000000000)
    netRate, err := strconv.ParseFloat(os.Getenv("NETRATE"), 64)
    if err != nil {
        log.Fatal("Unable to parse NET rate.")
    }


    // Instantiating go routines
    if cpuRate > 0 {
        go cpu(cpuRate)
    }

    if memInstance > 0 {
        go mem(memInstance)
    }

    if diskRate > 0 {
        go disk(diskRate)
    }

    // Need both netRate and netConc
    if netRate > 0 && netConc > 0 {
        go net(netConc, netRate)
    }

    // Wait forever
    var wg sync.WaitGroup
    wg.Add(1)
    wg.Wait()
}
