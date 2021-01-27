package main

// contractEventsTemplateContent contains the template string from contract_events.go.tmpl
var contractEventsTemplateContent = `{{- $contract := . -}}
{{- $logger := (print $contract.ShortVar "Logger") -}}
{{- range $i, $event := .Events }}

func ({{$contract.ShortVar}} *{{$contract.Class}}) {{$event.CapsName}}(
	opts *ethutil.SubscribeOpts,
	{{$event.IndexedFilterDeclarations -}}
) *{{$event.SubscriptionCapsName}} {
	if opts == nil {
		opts = new(ethutil.SubscribeOpts)
	}
	if opts.TickDuration == 0 {
		opts.TickDuration = ethutil.DefaultSubscribeOptsTickDuration
	}
	if opts.BlocksBack == 0 {
		opts.BlocksBack = ethutil.DefaultSubscribeOptsBlocksBack
	}

	return &{{$event.SubscriptionCapsName}}{
		{{$contract.ShortVar}},
		opts,
		{{$event.IndexedFilters}}
	}
}

type {{$event.SubscriptionCapsName}} struct {
	contract *{{$contract.Class}}
	opts *ethutil.SubscribeOpts
	{{$event.IndexedFilterFields -}}
}

type {{$contract.FullVar}}{{$event.CapsName}}Func func(
	{{$event.ParamDeclarations -}}
)

func ({{$event.SubscriptionShortVar}} *{{$event.SubscriptionCapsName}}) OnEvent(
	handler {{$contract.FullVar}}{{$event.CapsName}}Func,
) subscription.EventSubscription {
	eventChan := make(chan *abi.{{$contract.AbiClass}}{{$event.CapsName}})
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <- eventChan:
			    handler(
					{{$event.ParamExtractors}}
				)
			}
		}
	}()

	sub := {{$event.SubscriptionShortVar}}.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}

func ({{$event.SubscriptionShortVar}} *{{$event.SubscriptionCapsName}}) Pipe(
	sink chan *abi.{{$contract.AbiClass}}{{$event.CapsName}},
) subscription.EventSubscription {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker({{$event.SubscriptionShortVar}}.opts.TickDuration)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				lastBlock, err := {{$event.SubscriptionShortVar}}.contract.blockCounter.CurrentBlock()
				if err != nil {
					{{$logger}}.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock-{{$event.SubscriptionShortVar}}.opts.BlocksBack

				{{$logger}}.Infof(
					"subscription monitoring fetching past {{$event.CapsName}} events " +
					    "starting from block [%v]",
					fromBlock,
				)
				events, err := {{$event.SubscriptionShortVar}}.contract.Past{{$event.CapsName}}Events(
					fromBlock,
					nil,
					{{$event.IndexedFilterExtractors}}
				)
				if err != nil {
					{{$logger}}.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				{{$logger}}.Infof(
					"subscription monitoring fetched [%v] past {{$event.CapsName}} events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := {{$event.SubscriptionShortVar}}.contract.watch{{$event.CapsName}}(
		sink,
		{{$event.IndexedFilterExtractors}}
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}

func ({{$contract.ShortVar}} *{{$contract.Class}}) watch{{$event.CapsName}}(
	sink chan *abi.{{$contract.AbiClass}}{{$event.CapsName}},
	{{$event.IndexedFilterDeclarations -}}
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return {{$contract.ShortVar}}.contract.Watch{{$event.CapsName}}(
			&bind.WatchOpts{Context: ctx},
			sink,
			{{$event.IndexedFilters}}
		)
	}

	return ethutil.WithResubscription(
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

{{- end -}}`
