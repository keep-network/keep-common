package main

// contractEventsTemplateContent contains the template string from contract_events.go.tmpl
var contractEventsTemplateContent = `{{- $contract := . -}}
{{- $logger := (print $contract.ShortVar "Logger") -}}
{{- range $i, $event := .Events }}

func ({{$contract.ShortVar}} *{{$contract.Class}}) {{$event.CapsName}}(
	opts *ethutil.SubscribeOpts,
	{{$event.IndexedFilterDeclarations -}}
) *{{$event.CapsName}}Subscription {
	return &{{$event.CapsName}}Subscription{
		opts,
		{{$event.IndexedFilters}}
	}
}

type {{$event.CapsName}}Subscription struct {
	opts *ethutil.SubscribeOpts
	{{$event.IndexedFilterFields -}}
}

func ({{$event.ShortVar}}s *{{$event.CapsName}}Subscription) Pipe(
	chan *abi.{{$contract.AbiClass}}{{$event.CapsName}},
) {

}

type {{$contract.FullVar}}{{$event.CapsName}}Func func(
	{{$event.ParamDeclarations -}}
)

func ({{$event.ShortVar}}s *{{$event.CapsName}}Subscription) OnEvent(
	handler {{$contract.FullVar}}{{$event.CapsName}}Func,
) {

}

func ({{$contract.ShortVar}} *{{$contract.Class}}) Past{{$event.CapsName}}Events(
	startBlock uint64,
	endBlock *uint64,
	{{$event.IndexedFilterDeclarations -}}
) ([]*abi.{{$contract.AbiClass}}{{$event.CapsName}}, error){
	iterator, err := {{$contract.ShortVar}}.contract.Filter{{$event.CapsName}}(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		{{$event.IndexedFilters}}
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past {{$event.CapsName}} events: [%v]",
			err,
		)
	}

	events := make([]*abi.{{$contract.AbiClass}}{{$event.CapsName}}, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func ({{$contract.ShortVar}} *{{$contract.Class}}) Watch{{$event.CapsName}}(
	success {{$contract.FullVar}}{{$event.CapsName}}Func,
	{{$event.IndexedFilterDeclarations -}}
) (subscription.EventSubscription) {
	eventOccurred := make(chan *abi.{{$contract.AbiClass}}{{$event.CapsName}})

	ctx, cancel := context.WithCancel(context.Background())

	// TODO: Watch* function will soon accept channel as a parameter instead
	// of the callback. This loop will be eliminated then.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventOccurred:
				success(
                    {{$event.ParamExtractors}}
				)
			}
		}
	}()

	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return {{$contract.ShortVar}}.contract.Watch{{$event.CapsName}}(
			&bind.WatchOpts{Context: ctx},
			eventOccurred,
			{{$event.IndexedFilters}}
		)
	}

	sub := ethutil.WithResubscription(
		{{$contract.ShortVar}}SubscriptionBackoffMax,
		subscribeFn,
		{{$contract.ShortVar}}SubscriptionAlertThreshold,
		func(elapsed time.Duration) {
			{{$logger}}.Errorf(
					"subscription to event {{$event.CapsName}} had to be "+
						"retried [%v] since the last attempt; please inspect "+
						"Ethereum client connectivity",
					elapsed,
				)
		},
		func(err error) {
			{{$logger}}.Errorf(
					"subscription to event {{$event.CapsName}} failed "+
						"with error: [%v]; resubscription attempt will be "+
						"performed",
					err,
				)
		},
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}

{{- end -}}`
