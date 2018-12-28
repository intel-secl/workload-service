package setup

type SetupServer struct{}

func (ss SetupServer) Setup() error {
	// check if the PORT variable is set
	port := strconv.ParseInt(os.Getenv("WLS_PORTNUM"))
	return nil
}

func (ss SetupServer) Validate() error {
	return nil
	// validate that the port variable is not the zero value of its type
}
