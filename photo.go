package main

import "github.com/veandco/go-sdl2/sdl"
import "github.com/veandco/go-sdl2/sdl_image"

import "fmt"
import "os"
import "net/http"
import "io/ioutil"
import "encoding/xml"
import "unsafe"
import "net/url"
const winWidth = 1920
const winHeight= 1080

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
	api_key string
}

func (flickr *Flickr) gerUrl(method string, urlParams url.Values) string {
	var Url *url.URL

    Url, _ = url.Parse("https://api.flickr.com/services/rest/")

	urlParams.Add("method", method)
	urlParams.Add("api_key", flickr.api_key)

	Url.RawQuery = urlParams.Encode()
	return Url.String()
}
/*
func wget(chan data byte[], url string, args ...interface{}){
	response, err := http.Get(fmt.Sprintf(url,args))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load list: %s\n", err)
		return
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		   os.Exit(1)
	}
	data <- contents
}
*/

func wget(data chan []byte, url string){
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load list: %s\n", err)
		return
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		   os.Exit(1)
	}
	data <- contents
}

func main() {
	flickr := Flickr{"d23b3c30a27e62f70f3cf18b25d86a55"}
	data := make(chan []byte)

	go wget(data, flickr.gerUrl("flickr.people.getPublicPhotos", url.Values{"user_id": {"94969330@N02"}, "per_page" : {"10"}}))

	var rsp PhotoRsp

	if err := xml.Unmarshal(<- data, &rsp); 
	err != nil {
		panic(err)
	}

	for index,element := range rsp.Photos.Photo {
		fmt.Printf("[%d] %s \n", index, element.Title)
	}

	var s int
	_, _ = fmt.Scanf("%d", &s)

	go wget(data, flickr.gerUrl("flickr.photos.getSizes", url.Values{"photo_id": {rsp.Photos.Photo[s].ID} }))

	var srsp SizeRsp

	

	if err := xml.Unmarshal(<- data, &srsp); 
	err != nil {
		panic(err)
	}
	//fmt.Println(uri)

	for index,element := range srsp.Sizes.Size {
		fmt.Printf("[%d] %s \n", index, element.Source)
	}

	var i int
	_, _ = fmt.Scanf("%d", &i)



	fmt.Printf("Loadig: %s \n",srsp.Sizes.Size[i].Source)

	go wget(data, srsp.Sizes.Size[i].Source)


  
	fmt.Printf("Photo: %s\n",srsp.Sizes.Size[i].Source)
	
	contents := <- data

	rwops := sdl.RWFromMem(unsafe.Pointer(&contents[0]), len(contents))
	defer rwops.RWclose()
	surf, err := img.Load_RW(rwops, false)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create surface: %s\n", err)
		return
	}

	sdl.Init(sdl.INIT_EVERYTHING)
	window, err := sdl.CreateWindow("Images", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN | sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return
	}
	defer renderer.Destroy()

/*	image,_ := img.Load("../flicker/images/13338279973_5945cdae28_z.jpg")

	texture, err := renderer.CreateTextureFromSurface(image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return
//	}
	defer texture.Destroy()
*/

	//src := sdl.Rect{0, 0, 1024, 1024}
	var x int32
	var y int32
	
	var coef int32 = 1
	if surf.W > winWidth {
		coef = (surf.W/winWidth)
	}

	if surf.H > winHeight {
		coef = (surf.H/winHeight)
	}
	x = surf.W/coef;
	y = surf.H/coef;
	dst := sdl.Rect{int32((winWidth-x)/2) ,int32((winHeight-y)/2) , surf.W, surf.H}

	renderer.Clear()
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.FillRect(&sdl.Rect{0, 0, int32(winWidth), int32(winHeight)})
//	renderer.Copy(texture, &src, &dst)
	renderer.Present()
	fmt.Printf("img size: (%d, %d)\n", surf.W, surf.H)

	texture, err := renderer.CreateTextureFromSurface(surf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return
	}
	defer texture.Destroy()
	
	
	renderer.Clear()
	renderer.SetDrawColor(0, 0, 0, 255)

	renderer.FillRect(&sdl.Rect{0, 0, int32(winWidth), int32(winHeight)})

	renderer.Copy(texture, nil, &dst)
	renderer.Present()
	
	fmt.Scanln()
	sdl.Quit()
}