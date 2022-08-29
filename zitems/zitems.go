package zitems

type Item struct {
	ResourceID string
	Name       string
	DataPtr    any
	UpdateFunc func(newData any) `json:"-"`
	GetFunc    func()            `json:"-"`
}

type ZItemsCalls struct{}

var (
	AllItems []Item
	Calls    = new(ZItemsCalls)
)

func FindItem(resourceID string) (*Item, int) {
	for i := range AllItems {
		if AllItems[i].ResourceID == resourceID {
			return &AllItems[i], i
		}
	}
	return nil, -1
}

func RegisterItem(data any, resourceID, name string) *Item {
	item := Item{DataPtr: data, Name: name, ResourceID: resourceID}
	AllItems = append(AllItems, item)
	return &AllItems[len(AllItems)-1]
}

