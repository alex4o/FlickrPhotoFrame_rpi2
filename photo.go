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
import "runtime"

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

func (flickr *Flickr) listPhotos(user_id string, per_page string, res chan *PhotoRsp ){
	data := make(chan []byte)

	go wget(data, flickr.gerUrl("flickr.people.getPublicPhotos", url.Values{"user_id": { user_id }, "per_page" : {per_page}}))

	var rsp PhotoRsp

	if err := xml.Unmarshal(<- data, &rsp); 
	err != nil {
		panic(err)
	}

	res <- &rsp
}


func (flickr *Flickr) getUrls(in chan *PhotoRsp, out chan string){
	rsp := <- in
	//surfs := make(chan *sdl.Surface)
	data := make(chan []byte)

	for _,photo := range rsp.Photos.Photo {
		fmt.Printf("Loadig:%s \n", photo.Title)

		go wget(data, flickr.gerUrl("flickr.photos.getSizes", url.Values{"photo_id": { photo.ID }}))

		var srsp SizeRsp
		if err := xml.Unmarshal(<- data, &srsp); 
		err != nil {
			panic(err)
		}
		//fmt.Println(uri)
		out <- srsp.Sizes.Size[len(srsp.Sizes.Size)-1].Source

		
	}	
}

func loadPhotos(in chan string){
	sur := make(chan *sdl.Surface)
	//dst := make(chan *sdl.Rect)
	fmt.Printf("Loading photos")

	for url := range in {
		fmt.Printf("Loadig: %s \n",url)
		go loadPhoto(url, sur)
		fmt.Printf("Loaded: %s\n",url)
		
		surf := <- sur

		go func() {
			var x int32
			var y int32
			
			var coef int32 = 1
			if surf.W > winWidth {
				coef = (surf.W/winWidth)
			}

			if surf.H > winHeight {
				coef = (surf.H/winHeight)
			}
			x = surf.W/coef
			y = surf.H/coef
			fmt.Printf("img size: (%d, %d)\n", surf.W, surf.H)
			
			go func() {
				dst := &sdl.Rect{int32((winWidth-x)/2) ,int32((winHeight-y)/2) , surf.W, surf.H}
				


				do(func(r *sdl.Renderer){

					texture, err := r.CreateTextureFromSurface(surf)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
						return
					}
					defer texture.Destroy()

					r.Copy(texture, nil, dst)
				})
			}()
		}()

	}
}

func loadPhoto(url string ,out chan *sdl.Surface){
	data := make(chan []byte)

	fmt.Printf("Downloading: %s \n",url)
	go wget(data, url)
	
	contents := <- data

	rwops := sdl.RWFromMem(unsafe.Pointer(&contents[0]), len(contents))
	defer rwops.RWclose()
	surf, err := img.Load_RW(rwops, false)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create surface: %s\n", err)
		return
	}
	out <- surf
}

var mainfunc = make(chan func(*sdl.Renderer))

func do(f func(*sdl.Renderer)) {
	done := make(chan bool, 1)
	mainfunc <- func(r *sdl.Renderer) {
		r.Clear()
		f(r)
		r.Present()
		done <- true
	}
	<-done
}

func main() {
	runtime.LockOSThread()
	flickr := Flickr{"d23b3c30a27e62f70f3cf18b25d86a55"}
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

	res := make(chan *PhotoRsp)
	urls := make(chan string)

	go flickr.listPhotos("94969330@N02", "10", res)
	go flickr.getUrls(res, urls)

	go loadPhotos(urls)

	renderer.Clear()
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.FillRect(&sdl.Rect{0, 0, int32(winWidth), int32(winHeight)})
//	renderer.Copy(texture, &src, &dst)
	renderer.Present()
	
	for f := range mainfunc {
		f(renderer)
	} 

	//renderer.Copy(texture, nil, &dst)
	renderer.Present()
	
	fmt.Scanln()
	sdl.Quit()
}