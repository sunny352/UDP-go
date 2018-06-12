package Safe

import "github.com/sunny352/UDP-go/Utils/SLog"

func Go(function func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				SLog.E(err)
			}
		}()
		function()
	}()
}
