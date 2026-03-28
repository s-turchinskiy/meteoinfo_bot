package restyreceiver

import "github.com/go-resty/resty/v2"

type RestyReceiver struct {
	client *resty.Client
}

func NewRestyReceiver() *RestyReceiver {
	return &RestyReceiver{
		client: resty.New(),
	}
}

func (r *RestyReceiver) Receive(url string) (data []byte, err error) {
	resp, err := r.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}
