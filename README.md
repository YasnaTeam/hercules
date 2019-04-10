# Hercules :muscle: :trident:

Hercules is a concurrent download library for `GO` with a simple interface!

## Usages :loudspeaker:
Import it from `github.com/YasnaTeam/hercules` header, pass the file address, save path and the number of concurrent workers to the `Get` method and that's done!

## Samples :art:
These codes are tested under this environment:
```bash
yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ go version                                   
go version go1.10.4 linux/amd64

yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ uname -r
4.15.0-46-generic
```

### :one: Simple Download :pizza:
* :scroll: Code:
```go
package main  
  
import (  
	"fmt"
	"github.com/YasnaTeam/hercules"
)  
  
func main() {  
	elapsed, err := hercules.Get(  
		"https://upload.wikimedia.org/wikipedia/commons/e/e6/Hazy_Crazy_Sunrise.jpg",  
		"/tmp/sunrise.jpg",
		5,  
	)  
	if err != nil {  
		panic(err)  
	}  
  
	fmt.Printf("Total download time: %s\n", elapsed)  
}
```

* :chart_with_upwards_trend: Run:
```bash
yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ go run main.go
Total download time: 54.55608664s

yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ ll /tmp/sunrise.jpg
-rw-r--r-- 1 yasna yasna 12M Apr 10 12:21 /tmp/sunrise.jpg

yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ # due to the downloading time, it seems today the SSL handshakes has a serious problem, thanks to our big brother :/
```

### :two: Advanced usage of `hercules` methods :beers:
* :scroll: Code:
```go
package main  
  
import (  
	"fmt"  
	"os"  
	"github.com/YasnaTeam/hercules" 
	"github.com/sirupsen/logrus"
)

func main() {  
	// write only flag (os.O_WRONLY, or a flag with this privileges) is crucial on creating a file pointer
	fp, err := os.OpenFile("/tmp/sunrise.jpg", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)  
	if err != nil {  
		panic(err)  
	}
	defer fp.Close()  
  
	h, err := hercules.New(  
	      "https://upload.wikimedia.org/wikipedia/commons/e/e6/Hazy_Crazy_Sunrise.jpg",
	      fp,  
	      5,  
	)  
	if err != nil {  
		panic(err)  
	}  
  
	// set an external logger
	log := logrus.New()  
	log.Level = logrus.DebugLevel  
	h.SetLogger(log)  
  
	// fetch file size and ensure that the server supports multi-part downloading
	if err := h.Preload(); err != nil {  
		panic(err)  
	}  
  
	h.StartAll()
  
	h.Wait()  
	fmt.Printf("Total download time: %s\n", h.Elapsed())
}
```
* :chart_with_upwards_trend: Run:
```bash
yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ go run main.go
INFO[0001] Total size of file is 11915661B              
INFO[0001] A new worker started...                       
INFO[0001] A new worker started...                       
INFO[0001] A new worker started...                       
INFO[0001] A new worker started...                       
INFO[0001] A new worker started...                       
INFO[0001] Wait for finishing download...               
INFO[0001] Start downloading part #0...                 
INFO[0001] Start downloading part #3...                 
INFO[0001] Start downloading part #2...                 
INFO[0001] Start downloading part #1...                 
INFO[0001] Start downloading part #4...                 
INFO[0001] Writing offset 0 to the disk...              
INFO[0001] Writing offset 7149396 to the disk...        
INFO[0001] Writing offset 2383132 to the disk...        
INFO[0001] Writing offset 4766264 to the disk...        
INFO[0002] Writing offset 9532528 to the disk...        
INFO[0019] Part #4 is done (18.230438277s).             
INFO[0019] End of downloading part #4 (18.23035141s)... 
INFO[0021] Part #0 is done (20.054390877s).             
INFO[0021] End of downloading part #0 (20.05481599s)... 
INFO[0023] Part #3 is done (22.30961579s).              
INFO[0023] End of downloading part #3 (22.309480659s)... 
INFO[0024] Part #1 is done (22.931328433s).             
INFO[0024] End of downloading part #1 (22.931287887s)... 
INFO[0024] Part #2 is done (23.357667822s).             
INFO[0024] End of downloading part #2 (23.357594625s)... 
Total download time: 23.357789185s

yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ ll /tmp/sunrise.jpg
-rw-r--r-- 1 yasna yasna 12M Apr 10 12:23 /tmp/sunrise.jpg

yasna@freedom:~/Go/src/github.com/meysampg/testDownloader 
$ # in life, there are wounds that slowly annihilate the soul in isolation and our big brother is just one of them (probability the biggest) X(
``` 


## Get It! :tada:
The library can be fetched in the normal manner which you install other libraries with `github.com/YasnaTeam/hercules` :see_no_evil:.

 * `get`: `go get github.com/YasnaTeam/hercules`
 * `dep`: `dep ensure -add github.com/YasnaTeam/hercules`

## ToDo :thought_balloon:
 * Make the error channel on `Get` method more usable (or even usable).
 * Add `context` support to the downloader
 * Make `logger` support more useful and flexible

------------
Made with :heart: in [YasnaTeam](https://yasna.team).
