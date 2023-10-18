package main

/*
func TestWS(t *testing.T) {

	u, err := getWsURL("token", "homeid")
	assert.NoError(t, err)
	fmt.Println("url is", u)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	err = reconnectWS(ctx, u, "token", "homeid", func(data *DataPayload) {
		fmt.Println("the data is")
		spew.Dump(data)
	})
	assert.NoError(t, err)
}
*/
