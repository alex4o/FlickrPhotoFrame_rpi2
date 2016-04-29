package flickr

import "encoding/xml"
import "fmt"
import "../web"
import "net/url"


type PhotoRsp struct {
	Photos struct {
		Page    int `xml:"page,attr"`
		Pages   int `xml:"pages,attr"`
		Perpage int `xml:"perpage,attr"`
		Total string `xml:"total,attr"`
		Photo   []struct {
			Farm     int    `xml:"farm,attr"`
			ID       string `xml:"id,attr"`
			Isfamily int    `xml:"isfamily,attr"`
			Isfriend int    `xml:"isfriend,attr"`
			Ispublic int    `xml:"ispublic,attr"`
			Owner    string `xml:"owner,attr"`
			Secret   string `xml:"secret,attr"`
			Server   string `xml:"server,attr"`
			Title    string `xml:"title,attr"`
		} `xml:"photo"`
	} `xml:"photos"`
	Stat string `xml:"stat,attr"`
}

type SizeRsp struct {
	XMLName xml.Name `xml:"rsp"`
	Sizes struct {
		Canblog     int `xml:"canblog,attr"`
		Candownload int `xml:"candownload,attr"`
		Canprint    int `xml:"canprint,attr"`
		Size        []struct {
			Height int    `xml:"height,attr"`
			Label  string `xml:"label,attr"`
			Media  string `xml:"media,attr"`
			Source string `xml:"source,attr"`
			URL    string `xml:"url,attr"`
			Width  int    `xml:"width,attr"`
		} `xml:"size"`
	} `xml:"sizes"`
	Stat string `xml:"stat,attr"`
}

type Flickr struct {
	Api_key string
}

func (flickr *Flickr) GetUrl(method string, urlParams url.Values) string {
	var Url *url.URL

	Url, _ = url.Parse("https://api.flickr.com/services/rest/")

	urlParams.Add("method", method)
	urlParams.Add("api_key", flickr.Api_key)

	Url.RawQuery = urlParams.Encode()
	return Url.String()
}

func (flickr *Flickr) ListPhotos(user_id string, per_page string, page int32, res chan *PhotoRsp ){
	data := make(chan []byte, 1)

	go web.Get(data, flickr.GetUrl("flickr.people.getPublicPhotos", url.Values{"user_id": { user_id }, "per_page" : {per_page}, "page": {fmt.Sprintf("%d",page)}}))

	var rsp PhotoRsp

	if err := xml.Unmarshal(<- data, &rsp);
	err != nil {
		panic(err)
	}

	res <- &rsp
}

func (flickr *Flickr) GetFavs(user_id string, per_page string, page int32, res chan *PhotoRsp){
	data := make(chan []byte, 1)

	go web.Get(data, flickr.GetUrl("flickr.favorites.getPublicList", url.Values{"user_id": { user_id }, "per_page" : {per_page}, "page": {fmt.Sprintf("%d",page)}}))

	var rsp PhotoRsp

	if err := xml.Unmarshal(<- data, &rsp);
	err != nil {
		panic(err)
	}

	res <- &rsp
}

func (flickr *Flickr) GetUrls(in chan *PhotoRsp, out chan string){
	rsp := <- in
	//surfs := make(chan *sdl.Surface)
	data := make(chan []byte)

	for _,photo := range rsp.Photos.Photo {
		fmt.Printf("Loadig:%s \n", photo.Title)

		go web.Get(data, flickr.GetUrl("flickr.photos.getSizes", url.Values{"photo_id": { photo.ID }}))

		var srsp SizeRsp
		if err := xml.Unmarshal(<- data, &srsp);
		err != nil {
			panic(err)
		}
		//fmt.Println(uri)
		var size int = 1
		if len(srsp.Sizes.Size) >= 11 {
			size = 10
		}else{
			size = len(srsp.Sizes.Size)-1
		}

		out <- srsp.Sizes.Size[size].Source


	}
}
