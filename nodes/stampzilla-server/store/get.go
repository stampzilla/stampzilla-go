package store

func (store *Store) Get(area string) (interface{}, bool) {
	switch area {
	case "devices":
		return store.GetDevices(), true
	case "connections":
		return store.GetConnections(), true
	case "nodes":
		return store.GetNodes(), true
	case "certificates":
		return store.GetCertificates(), true
	case "requests":
		return store.GetRequests(), true
	case "rules":
		return store.GetRules(), true
	case "savedstates":
		return store.GetSavedStates(), true
	case "schedules":
		return store.GetScheduledTasks(), true
	case "server":
		return store.GetServerStateAsJson(), true
	case "destinations":
		return store.GetDestinations(), true
	case "senders":
		return store.GetSenders(), true
	case "persons":
		return store.GetPersons(), true
	case "cloud":
		return store.GetCloud(), true
	}
	return nil, false
}
