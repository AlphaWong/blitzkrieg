Blitzkrieg
==========
Blitzkrieg is a refactoring of [Dave Cheney](https://github.com/dave)'s work on [blast](https://github.com/dave/blast/).

Blitzkrieg builds solid foundation for building custom load testers with custom logic and behaviour yet with the 
statistical foundation provided in [blast](https://github.com/dave/blast/), for extensive details on the performance
of a API target. 

 * Blitzkrieg makes API requests at a fixed rate.
 * The number of concurrent workers is configurable.
 * The worker API allows custom protocols or logic for how a target get's tested
 * Blitzkrieg builds a solid foundation for custom load testers.

 ## From source
 
 ```
 go get -u github.com/gokit/blitzkrieg
 ```

 Status
 ======

 Blitzkrieg prints a summary every ten seconds. While Blitzkrieg is running, you can hit enter for an updated
 summary, or enter a number to change the sending rate. Each time you change the rate a new column
 of metrics is created. If the worker returns a field named `status` in it's response, the values
 are summarised as rows.

 Here's an example of the output:

 ```
 Metrics
 =======
 Concurrency:      1999 / 2000 workers in use

 Desired rate:     (all)        10000        1000         100
 Actual rate:      2112         5354         989          100
 Avg concurrency:  1733         1976         367          37
 Duration:         00:40        00:12        00:14        00:12

 Total
 -----
 Started:          84525        69004        14249        1272
 Finished:         82525        67004        14249        1272
 Mean:             376.0 ms     374.8 ms     379.3 ms     377.9 ms
 95th:             491.1 ms     488.1 ms     488.2 ms     489.6 ms

 200
 ---
 Count:            79208 (96%)  64320 (96%)  13663 (96%)  1225 (96%)
 Mean:             376.2 ms     381.9 ms     374.7 ms     378.1 ms
 95th:             487.6 ms     489.0 ms     487.2 ms     490.5 ms

 404
 ---
 Count:            2467 (3%)    2002 (3%)    430 (3%)     35 (3%)
 Mean:             371.4 ms     371.0 ms     377.2 ms     358.9 ms
 95th:             487.1 ms     487.1 ms     486.0 ms     480.4 ms

 500
 ---
 Count:            853 (1%)     685 (1%)     156 (1%)     12 (1%)
 Mean:             371.2 ms     370.4 ms     374.5 ms     374.3 ms
 95th:             487.6 ms     487.1 ms     488.2 ms     466.3 ms

 Current rate is 10000 requests / second. Enter a new rate or press enter to view status.

 Rate?
 ```



Worker API
==========

Blitzkrieg Worker interface allows Blitzkrieg to support your custom load testing strategies. 

The workers used by blitskrieg for your load test is generated by you when you provide 
the following function, which returns new instances for use in concurrently load testing 
their internal target. 

*In blitzkrieg, you handle the target you wish to target in your worker and the data they 
will be using for their tests. Remember blitzkrieg is a foundation, it collects the stats
for you, so you just extend it for custom load testing setup.*

```go
func sampleWorker() blitzkriege.Worker {
	return &LalaWorker{}
}
```


## Worker Examples
See Worker interface implementation examples below:

### Single Request Worker  

This is where a single request is to be made to hit at our target as defined by us. This worker
will be called to repeatedly prepare it's payload and then we measure how long it takes for it 
to make it's request and get it's response.

```go
type LalaServiceConfig struct{
	MainServiceAddr string
	MainServicePort int
	TlsCert *tls.Cert
}

type LalaWorker struct {}

// Send will contain all necessary call or calls require for your load testing
// You can use the WorkerContext.FromContext method to create a child context to 
// detail the response, status and error that occurs from that sub-request, this then
// allows us follow that tree to create comprehensive statistics for your load test.
func (e *LalaWorker) Send(ctx context.Context,  workerContext *WorkerContext) error {
	
	// Call target service and record response, err and status.
	resStatus, response, err := callMainService(ctx)
	
	workerContext.SetResponse(resStatus, Payload{ Body: response }, err)
	return err
}

// Prepare Start should be where you load the sample data you wish to test your worker with.
// You might need to implement some means of selectively or randomly loading different
// data for your load tests here, as Blitzkrieg won't handle that for you.
//
// This is called on every time a worker is to be executed, so you get the chance
// to prepare the payload data, headers and parameters you want in a request.
func (e *LalaWorker) Prepare(ctx context.Context) (*WorkerContext, error) {
	// Load some data from disk or some remote service
	var customBody, err = LoadData()
	if err != nil {
		return nil, err
	}
	
	// You can add custom parameters and headers or load this all from 
	// some custom encoded file (e.g in JSON).
	var customParameters = map[string]string{"user_id": "some_id"}
	var customHeader = map[string][]string{
		"X-Record-Meta": []string{"raf-4334", "xaf-rt"},
	}
	
	
	// create custom request payload.
	var payload = blitzkrieg.Payload{
		Body: customBody, 
		Params: customParameters, 
		Headers: customHeader,
	}
	
	// create or load some custom meta data or config for worker.
	var serviceMeta = LalaServiceConfig{}
	
	return blitskrieg.NewWorkerContext("raf-api-test", payload, serviceMeta)
}

// You handle some base initialization logic you wish to be done before worker use.
//
// Remember Blitskrieg will create multiple versions of this worker with the 
// register WorkerFunc, so don't cause race conditions.
func (e *LalaWorker) Start(ctx context.Context) error {
	
	return nil
}

// You handle whatever cleanup you wish to be done for this worker.
//
// Remember Blitskrieg will create multiple versions of this worker with the 
// register WorkerFunc, so don't cause race conditions.
func (e *LalaWorker) Stop(ctx context.Context) error {
	// do something....
}
```

### Group/Sequence Request Worker

This is where your load test spans multiple requests, where each makes up the single operation 
you wish to validate it's behaviour, using the `WorkerContext.FromContext` which branches of 
child WorkerContext, we can ensure the worker and this sub request data are measured and 
and aggregated.

```go
type LalaServiceConfig struct{
	MainServiceAddr string
	MainServicePort int
	TlsCert *tls.Cert
}

type LalaWorker struct {}

// Send will contain all necessary call or calls require for your load testing
// You can use the WorkerContext.FromContext method to create a child context to 
// detail the response, status and error that occurs from that sub-request, this then
// allows us follow that tree to create comprehensive statistics for your load test.
func (e *LalaWorker) Send(ctx context.Context,  workerContext *WorkerContext) error {
	
	// Make request 1 to first API endpoint using child context.
	firstWorkerContext  := workerContext.FromContext("request-1", Payload{})
	if err := callFirstService(ctx, secondWorkerContext); err != nil {
		return err
	}
	
	// Make request 2 to next API endpoint in series
	secondWorkerContext  := workerContext.FromContext("request-2", Payload{})
	if err := callSecondService(ctx, secondWorkerContext); err != nil {
		return err
	}
	
	// Make main request with worker context
	resStatus, response, err := callMainService(ctx)
	 
	workerContext.SetResponse(resStatus, Payload{ Body: response }, err)
	return err
}

// Prepare Start should be where you load the sample data you wish to test your worker with.
// You might need to implement some means of selectively or randomly loading different
// data for your load tests here, as Blitzkrieg won't handle that for you.
//
// This is called on every time a worker is to be executed, so you get the chance
// to prepare the payload data, headers and parameters you want in a request.
func (e *LalaWorker) Prepare(ctx context.Context) (*WorkerContext, error) {
	// Load some data from disk or some remote service
	var customBody, err = LoadData()
	if err != nil {
		return nil, err
	}
	
	// You can add custom parameters and headers or load this all from 
	// some custom encoded file (e.g in JSON).
	var customParameters = map[string]string{"user_id": "some_id"}
	var customHeader = map[string][]string{
		"X-Record-Meta": []string{"raf-4334", "xaf-rt"},
	}
	
	
	// create custom request payload.
	var payload = blitzkrieg.Payload{
		Body: customBody, 
		Params: customParameters, 
		Headers: customHeader,
	}
	
	// create or load some custom meta data or config for worker.
	var serviceMeta = LalaServiceConfig{}
	
	return blitskrieg.NewWorkerContext("raf-api-test", payload, serviceMeta)
}

// You handle some base initialization logic you wish to be done before worker use.
//
// Remember Blitskrieg will create multiple versions of this worker with the 
// register WorkerFunc, so don't cause race conditions.
func (e *LalaWorker) Start(ctx context.Context) error {
	
	return nil
}

// You handle whatever cleanup you wish to be done for this worker.
//
// Remember Blitskrieg will create multiple versions of this worker with the 
// register WorkerFunc, so don't cause race conditions.
func (e *LalaWorker) Stop(ctx context.Context) error {
	// do something....
}
```

### More Examples

```go
ctx, cancel := context.WithCancel(context.Background())
b := blitzkrieg.New(ctx, cancel)
defer b.Exit()

b.SetWorker(func() blitzkrieg.Worker {
	return &blitzkrieg.FunctionWorker{
		SendFunc: func(ctx context.Context, workerCtx *blistskrieg.WorkerContext) error {
			workerCtx.SetStatus(blitzkrieg.Stringify(200))
			return nil
		},
		Start: func(ctx context.Context) (*blitzkireg.WorkerContext, error){
			var payload blitskrieg.Payload
			payload.Headers = []string{"header"}
			payload.SetData(strings.NewReader("foo\nbar"))
			return blitzkrieg.NewWorkerContext("some-text", payload, nil)
		}
	}
})

stats, err := b.Start(ctx, blitskrieg.Config{
	Rate: 1000,
})

if err != nil {
	fmt.Println(err.Error())
	return
}

fmt.Printf("Success == 2: %v\n", stats.All.Summary.Success == 2)
fmt.Printf("Fail == 0: %v", stats.All.Summary.Fail == 0)
// Output:
// Success == 2: true
// Fail == 0: true
```

```go
ctx, cancel := context.WithCancel(context.Background())
b := blitzkrieg.New(ctx, cancel)
defer b.Exit()

b.SetWorker(func() blitzkrieg.Worker {
	return &blitzkrieg.FunctionWorker{
		SendFunc: func(ctx context.Context, workerCtx *blistskrieg.WorkerContext) error {
			workerCtx.SetStatus(blitzkrieg.Stringify(200))
			return nil
		},
		Start: func(ctx context.Context) (*blitzkireg.WorkerContext, error){
			var payload blitskrieg.Payload
			payload.Headers = []string{"header"}
			payload.SetData(strings.NewReader("foo\nbar"))
			return blitzkrieg.NewWorkerContext("some-text", payload, nil)
		}
	}
})

wg := &sync.WaitGroup{}
wg.Add(1)

go func() {
	stats, err := b.Start(ctx, blitskrieg.Config{
		Rate: 1000,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Success > 10: %v\n", stats.All.Summary.Success > 10)
	fmt.Printf("Fail == 0: %v", stats.All.Summary.Fail == 0)
	wg.Done()
}()

<-time.After(time.Millisecond * 100)
b.Exit()

wg.Wait()

// Output:
// Success > 10: true
// Fail == 0: true
```
 
