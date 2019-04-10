package hercules

import "os"

// Get download a given address with workers and save it on the disk
func Get(addr, savePath string, workerCount int) (string, error) {
	fp, err := os.OpenFile(savePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	h, err := New(addr, fp, workerCount)
	if err != nil {
		return "", err
	}

	if err := h.Preload(); err != nil {
		return "", err
	}

	errChan := h.StartAll() // how to use this channel normally? :S
	defer close(errChan)
	go checkErrs(errChan)

	h.Wait()

	return h.Elapsed(), nil
}

func checkErrs(err chan error) {
	select {
		case e := <- err:
			panic(e)

	default:
		// :soOo0t
	}
}
