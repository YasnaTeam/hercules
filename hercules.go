package hercules

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type logger interface {
	Print(v ...interface{})
	Error(v ...interface{})
}

// Part shows a part of bytes which must be downloaded
type Part struct {
	Start int64 // start byte of download
	End   int64 // end byte of download
}

// Hercules is who can download anything, anytime :D
type Hercules struct {
	Addr string   // download address
	File *os.File // save file pointer

	fs        int64          // file size
	parts     []Part         // download parts
	wg        sync.WaitGroup // wait group
	logger    logger         // logger
	startTime time.Time      // time of starting download
}

// New returns a new instance of Hercules.
func New(addr string, fp *os.File, numberOfWorker ...int) (*Hercules, error) {
	if fp == nil {
		return nil, errors.New("file pointer can't be nil")
	}

	var c int
	if 0 != len(numberOfWorker) {
		c = numberOfWorker[0]
	}

	return &Hercules{
		Addr: addr,
		File: fp,

		parts: make([]Part, c),
	}, nil
}

// Preload prepare Hercules to start download
func (h *Hercules) Preload() error {
	if len(h.parts) == 0 {
		h.error("for using Preload, you must set number of workers first")
		return errors.New("for using Preload, you must set number of workers first")
	}

	if err := h.FetchHeaders(); err != nil {
		return err
	}

	if err := h.GenerateParts(); err != nil {
		return err
	}

	return nil
}

// StartAll run all workers!
func (h *Hercules) StartAll() chan error {
	errChan := make(chan error, len(h.parts))

	h.startTime = time.Now()
	for i := range h.parts {
		h.Run()
		go func(i int) {
			errChan <- h.Start(i)
		}(i)
	}

	return errChan
}

// Start starts a worker to download his part
func (h *Hercules) Start(partNum int) error {
	if 0 == len(h.parts) || partNum >= len(h.parts) {
		h.error("the part number is greater than the capacity")
		return errors.New("the part number is greater than the capacity")
	}

	startTime := time.Now()
	h.log(fmt.Sprintf("Start downloading part #%d...", partNum))
	if err := h.getPart(partNum); err != nil {
		h.error(err)
		return err
	}
	h.log(fmt.Sprintf("End of downloading part #%d (%s)...", partNum, time.Since(startTime)))

	return nil
}

func (h *Hercules) Run() {
	h.log("A new worker started...")
	h.wg.Add(1)
}

// Wait waits until all workers finish their works
func (h *Hercules) Wait() {
	h.log("Wait for finishing download...")
	h.wg.Wait()
}

// Done announce that a worker is done
func (h *Hercules) Done(n int) {
	h.wg.Done()
	h.log(fmt.Sprintf("Part #%d is done (%s).", n, h.Elapsed()))
}

func (h *Hercules) Elapsed() string {
	return time.Since(h.startTime).String()
}

// AddPart appends a part of download to the queue
func (h *Hercules) AddPart(start, end int64) {
	h.log(fmt.Sprintf("A part has been added to the parts (%dB).", end-start))
	h.parts = append(h.parts, Part{Start: start, End: end})
}

// AddPartOn puts a part of download to a specified location
func (h *Hercules) AddPartOn(index int, start, end int64) error {
	if 0 == len(h.parts) || index >= len(h.parts) {
		h.error("the part number is greater than the capacity")
		return errors.New("the part number is greater than the capacity")
	}

	h.log(fmt.Sprintf("Part #%d has been added to the parts (%dB).", index, end-start))
	h.parts[index] = Part{
		Start: start,
		End:   end,
	}

	return nil
}

// SetWorkerNumber sets number of concurrent downloader
func (h *Hercules) SetWorkerNumber(n int) {
	h.parts = make([]Part, n)
}

func (h *Hercules) SetLogger(l logger) {
	h.logger = l
}

// FetchHeaders checks server for support of multi-part downloading and getting file size
func (h *Hercules) FetchHeaders() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.Addr, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if _, ok := res.Header["Accept-Ranges"]; !ok {
		h.error("server doesn't support multi-part download")
		return errors.New("server doesn't support multi-part download")
	}

	size, err := strconv.ParseInt(res.Header["Content-Length"][0], 10, 64)
	if err != nil {
		return err
	}

	h.log(fmt.Sprintf("Total size of file is %dB", size))
	h.fs = size

	return nil
}

// GenerateParts
func (h *Hercules) GenerateParts() error {
	if h.fs == 0 {
		h.error("could not generate parts for a zero-sized destination")
		return errors.New("could not generate parts for a zero-sized destination")
	}

	n := int64(len(h.parts))
	if n == 0 {
		h.error("could not generate parts for zero workers")
		return errors.New("could not generate parts for zero workers")
	}

	partSize := int64(h.fs / n)
	var part int64
	for part = 0; part < n; part++ {
		if part == n-1 { // last part should download the complete size
			h.parts[part] = Part{
				Start: part * partSize,
				End:   h.fs,
			}
		} else { // split file between to workers
			h.parts[part] = Part{
				Start: part * partSize,
				End:   (part+1)*partSize - 1,
			}
		}
	}

	return nil
}

func (h *Hercules) getPart(partNum int) error {
	if 0 == len(h.parts) || partNum >= len(h.parts) {
		return errors.New("the part number is greater than the capacity")
	}

	var client http.Client
	req, err := http.NewRequest("GET", h.Addr, nil)
	if err != nil {
		h.Done(4)

		return err
	}

	bytesRange := fmt.Sprintf("bytes=%d-%d", h.parts[partNum].Start, h.parts[partNum].End)
	req.Header.Add("Range", bytesRange)
	resp, err := client.Do(req)
	if err != nil {
		h.Done(4)

		return err
	}

	size, err := strconv.ParseInt(resp.Header["Content-Length"][0], 10, 64)
	if err != nil {
		h.Done(4)

		return err
	}

	var partSize int64
	if partNum == len(h.parts)-1 {
		partSize = h.parts[partNum].End - h.parts[partNum].Start
	} else {
		partSize = h.parts[partNum].End - h.parts[partNum].Start + 1
	}

	if size != partSize {
		h.Done(4)
		h.error(fmt.Sprintf("could not fetch part #%d, wants %dB, %dB given", partNum, partSize, size))

		return errors.New(fmt.Sprintf("could not fetch part #%d, wants %dB, %dB given", partNum, partSize, size))
	}

	if err := h.savePartOnDisk(resp.Body, partNum); err != nil {
		return err
	}

	return nil
}

func (h *Hercules) savePartOnDisk(body io.ReadCloser, n int) error {
	defer body.Close()
	defer h.Done(n) // release the sync lock

	buf := make([]byte, 4*1024)
	offset := h.parts[n].Start
	h.log(fmt.Sprintf("Writing offset %d to the disk...", offset))
	for {
		b, err := body.Read(buf)
		if err != nil {
			if "EOF" == err.Error() {
				break
			}

			return err
		}

		nw, err := h.File.WriteAt(buf[0:b], offset)
		if err != nil {
			h.error(fmt.Sprintf("an error on writing part #%d occurred: %s", n, err))
			return errors.New(fmt.Sprintf("an error on writing part #%d occurred: %s", n, err))
		}

		offset = int64(nw) + offset
	}

	return nil
}

func (h *Hercules) log(a interface{}) {
	if h.logger != nil {
		h.logger.Print(a)
	}
}

func (h *Hercules) error(a interface{}) {
	if h.logger != nil {
		h.logger.Error(a)
	}
}
