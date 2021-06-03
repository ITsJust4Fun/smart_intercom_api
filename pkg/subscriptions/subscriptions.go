package subscriptions

import (
	"smart_intercom_api/graph/model"
	"sync"
)

var VideoUpdatedObservers map[string]chan *model.Video
var VideoUpdatedMutex sync.Mutex

func Init() {
	VideoUpdatedObservers = map[string]chan *model.Video{}
}
