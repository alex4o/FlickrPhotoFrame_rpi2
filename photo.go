package main

import "github.com/veandco/go-sdl2/sdl"
import "github.com/veandco/go-sdl2/sdl_image"
import "time"
import "fmt"
import "os"

import "unsafe"
import "runtime"

import "./flickr"
import "./web"

const winWidth = 1920
const winHeight = 1080

func loadPhotos(in chan string){
	sur := make(chan *sdl.Surface, 2)
	//dst := make(chan *sdl.Rect)
	fmt.Printf("Loading photos")

	for url := range in {
		fmt.Printf("Loadig: %s \n", url)
		go loadPhoto(url, sur)
		fmt.Printf("Loaded: %s\n", url)
		surf := <-sur

		go func() {
			var x int32
			var y int32

			var coef float32 = 1
			var coef1 float32 = 1
			var coef2 float32 = 1

			if surf.W > winWidth {
				coef1 = (float32(surf.W) / float32(winWidth))
				fmt.Printf("coef[1]: %f\n", coef1)

			}

			if surf.H > winHeight {
				coef2 = (float32(surf.H) / float32(winHeight))
				fmt.Printf("coef[2]: %f\n", coef2)
			}

			if coef1 > coef2 {
				coef = coef1
			} else {
				coef = coef2
			}
			x = int32(float32(surf.W) / coef)
			y = int32(float32(surf.H) / coef)
			fmt.Printf("img size: (%d, %d)\ncomputed size: (%d, %d)\n", surf.W, surf.H, x, y)

			go func() {
				dst := &sdl.Rect{int32((winWidth - x) / 2), int32((winHeight - y) / 2), x, y}
				do(func(r *sdl.Renderer) {

					texture, err := r.CreateTextureFromSurface(surf)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
						return
					}
					defer texture.Destroy()

					r.Copy(texture, nil, dst)
					r.Present()
					surf.Free()
				})
			}()
		}()
	}
}

func loadPhoto(url string, out chan *sdl.Surface) {
	data := make(chan []byte, 1)

	fmt.Printf("Downloading: %s \n", url)
	go web.Get(data, url)

	contents := <-data

	rwops := sdl.RWFromMem(unsafe.Pointer(&contents[0]), len(contents))
	defer rwops.RWclose()
	surf, err := img.Load_RW(rwops, false)
	//rwops.FreeRW()
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
		done <- true
	}
	<-done
}

func main() {
	runtime.LockOSThread()
	api := flickr.Flickr{"d23b3c30a27e62f70f3cf18b25d86a55"}
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("Images", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN | sdl.WINDOW_OPENGL)
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

	sdl.ShowCursor(0)

	res := make(chan *flickr.PhotoRsp, 2)
	urls := make(chan string, 10)
	var page int32 = 1

	go api.ListPhotos("94969330@N02", "10", page, res)
	go api.GetUrls(res, urls)

	go loadPhotos(urls)

	renderer.Clear()
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.FillRect(&sdl.Rect{0, 0, int32(winWidth), int32(winHeight)})
	//	renderer.Copy(texture, &src, &dst)
	renderer.Present()
	var c int = 1
	for f := range mainfunc {
		f(renderer)
		time.Sleep(time.Second * 2)
		c++

		fmt.Printf("image [%d]\npage [%d]\n", c, page)
		if c == 10 {
			c = 1
			page += 1
			go api.ListPhotos("94969330@N02", "10", page, res)
			go api.GetUrls(res, urls)

		}
		// wrong to do that here
	}

	//renderer.Copy(texture, nil, &dst)
	renderer.Present()

	fmt.Scanln()
	sdl.Quit()
}
