package util

import "strconv"

type Call struct {
	Out   interface{}
	Err   error
	Index int
}

func RelObject(ids []string, streamsCount int, callFunc func(id string) (out interface{}, err error), parseFunc func(call Call)) {
	i := len(ids)
	if i == 0 {
		return
	}
	singlePartSize := 0
	if i >= streamsCount {
		singlePartSize = i / streamsCount
	} else {
		streamsCount = 1
		singlePartSize = i
	}
	n, m := 0, singlePartSize
	ch := make(chan Call, i)
	for j := 1; j <= streamsCount; j++ {
		if j == streamsCount {
			m = i
		}
		go func(part []string) {
			for _, id := range part {
				data := Call{}
				recId := id[:36]
				if data.Index, data.Err = strconv.Atoi(id[36:]); data.Err == nil {
					data.Out, data.Err = callFunc(recId)
				}
				select {
				case ch <- data:
				}
			}
		}(ids[n:m])
		n = m
		m += singlePartSize
	}
	for j := 1; j <= i; j++ {
		parseFunc(<-ch)
	}
	close(ch)
}
