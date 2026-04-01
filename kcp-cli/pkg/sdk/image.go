package sdk

import "fmt"

// imageClient 는 Glance (Image) API 클라이언트 구현체이다
type imageClient struct {
	c *Client
}

// NewImageClient 는 새로운 Image 클라이언트를 생성한다
func NewImageClient(c *Client) ImageClient {
	return &imageClient{c: c}
}

// ListImages 는 이미지 목록을 조회한다
func (ic *imageClient) ListImages(opts *ListOpts) (*ListResponse[Image], error) {
	path := "/image/images" + buildQuery(opts)
	var resp ListResponse[Image]
	if err := ic.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetImage 는 특정 이미지를 조회한다
func (ic *imageClient) GetImage(id string) (*Image, error) {
	var img Image
	if err := ic.c.Get(fmt.Sprintf("/image/images/%s", id), &img); err != nil {
		return nil, err
	}
	return &img, nil
}

// DeleteImage 는 이미지를 삭제한다
func (ic *imageClient) DeleteImage(id string) error {
	return ic.c.Delete(fmt.Sprintf("/image/images/%s", id))
}
