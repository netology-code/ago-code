package app

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
)

type Watcher struct {
	client      *api.Client
}

func NewWatcher(client *api.Client) *Watcher {
	return &Watcher{client: client}
}


func (w *Watcher) Watch(services []string) error {
	errs := make(chan error, 1)

	for _, service := range services {
		go func(s string) {
			plan, err := watch.Parse(map[string]interface{}{"type": "service", "service": s, "passingonly": true})
			if err != nil {
				errs <- err
				return
			}
			// можем подписываться на изменения
			plan.HybridHandler = func(val watch.BlockingParamVal, i interface{}) {
				if entries, ok := i.([]*api.ServiceEntry); ok {
					for _, entry := range entries {
						log.Printf("%#v", entry)
					}
				}
			}
			err = plan.RunWithClientAndHclog(w.client, nil)
			if err != nil {
				errs <- err
				return
			}
		}(service)
	}

	return <- errs
}
