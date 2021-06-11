// Copyright Sebastian Kind 2018

// radar <- NAME FINDEN ist ein Wrapper für pcmet, um Wetterradardaten
// direkt aus dem Desktop laden zu können.

// no gui brach

// remove gui stuff

// place this on server and download images convert the to webp reduce size lal

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var limit int = 8
var usernameStr, passwdStr, viewerStr, pathStr *string
var compressBool *bool
var compressLevel *int

var COMPRESSION_LEVEL = 20 // 1..100

func main() {
	// flagparsing
	usernameStr = flag.String("user", "", "pcmet Benutzername")
	passwdStr = flag.String("passwd", "", "pcmet Passwort")
	viewerStr = flag.String("viewer", "eom", "Programm für Bilbetrachtung auswählen")
	mapStr := flag.String("map", "de", "Karte die abgerufen werden soll")
	pathStr = flag.String("path", "", "Pfad wo Bilder gespeichert werden (leer falls default tmpdir")
	compressBool = flag.Bool("compress", false, "JPEG Compression, wenn nicht angeben werden PNGs gespeichert")
	compressLevel = flag.Int("clevel", 15, "JPEG Compression, compress wird automatisch gesetzt")

	flag.Parse() //wichtig
	if *usernameStr == "" || *viewerStr == "" {
		log.Fatalf("Username/Password missing please use radar -user NAME -passwd yourPASSWORD")
	}

	COMPRESSION_LEVEL = *compressLevel

	// los gehts

	pcmet(*mapStr)
}

// // systray GUI
// func onReady() {
// 	systray.SetIcon(Data)
// 	systray.SetTitle("Radar")
// 	systray.SetTooltip("DWD pc_met Wetterinformationen")

// 	// menu items
// 	mEuro := systray.AddMenuItem("Radar Europa", "Öffnet Regenradar in neuem Fenster")
// 	mDe := systray.AddMenuItem("Radar Deutschland", "Öffnet Regenradar in neuem Fenster")
// 	mDeN := systray.AddMenuItem("Radar Deutschland Nord", "Öffnet Regenradar in neuem Fenster")
// 	mDeS := systray.AddMenuItem("Radar Deutschland Süd", "Öffnet Regenradar in neuem Fenster")
// 	mAlpen := systray.AddMenuItem("Radar Alpen", "Öffnet Regenradar in neuem Fenster")
// 	systray.AddSeparator()
// 	mSatEuro := systray.AddMenuItem("IR RGB Sat Europa", "Öffnet Satellitenbilder  in neuem Fenster")
// 	mSatDE := systray.AddMenuItem("IR RGB Sat Deutschland", "Öffnet Satellitenbilder in neuem Fenster")
// 	systray.AddSeparator()
// 	mLimitToggle := systray.AddMenuItem("8 Bilder -> 35 auswählen", "Bilderlimit auswählen")

// 	mQuitOrig := systray.AddMenuItem("Beenden", "Radar beenden")

// 	go func() {
// 		<-mQuitOrig.ClickedCh
// 		fmt.Println("Requesting quit")
// 		systray.Quit()
// 		fmt.Println("Finished quitting")
// 	}()

// 	for {
// 		select {
// 		case <-mEuro.ClickedCh:
// 			pcmet("eu")
// 		case <-mDe.ClickedCh:
// 			pcmet("rx")
// 		case <-mDeS.ClickedCh:
// 			pcmet("rxs")
// 		case <-mDeN.ClickedCh:
// 			pcmet("rxn")
// 		case <-mAlpen.ClickedCh:
// 			pcmet("fa")
// 		case <-mSatEuro.ClickedCh:
// 			pcmet("ir_rgb_eu")
// 		case <-mSatDE.ClickedCh:
// 			pcmet("ir_rgb_mdl")
// 		case <-mLimitToggle.ClickedCh:
// 			if limit == 8 {
// 				limit = 35 // i think this was the max amount of images on the server
// 				mLimitToggle.SetTitle("35 Bilder -> 8 auswählen")
// 			} else {
// 				limit = 8
// 				mLimitToggle.SetTitle("8 Bilder -> 35 auswählen")
// 			}
// 		case <-mQuitOrig.ClickedCh:
// 			systray.Quit()
// 			fmt.Println("Quit now")
// 			return
// 		}
// 	}

// }

// Aus der Wetterapp übernommen
//
func pcmet(region string) {

	path := ""
	switch region {
	case "fa", "eu", "rx", "rxs", "rxn", "rxm":
		path = "https://www.flugwetter.de/fw/scripts/getimg.php?src=rad/"
	default:
		path = "https://www.flugwetter.de/fw/scripts/getimg.php?src=sat/"
	}
	//	path := "/fw/scripts/getimg.php?src=rad/"
	dwdRadar := "https://www.flugwetter.de/fw/bilder/rad/index.htm?type=" + region

	// fixme dieses Teil mit waitgroup parallel ausführen lassen, hier wartet man auf den DWD Server

	imageString := getImageListString(authLoad(dwdRadar))

	// fixme

	images := makeSlice(imageString)

	tempdir := ""
	if *pathStr == "" {
		tempdir, err := ioutil.TempDir("", "radarImages")
		check(err)
		defer os.RemoveAll(tempdir)
	} else {
		tempdir = *pathStr
	}

	fmt.Println("Downloading images")
	for i, image := range images {
		//go mySpinner()
		pngData := []byte(authLoad(path + image))
		//fmt.Println(path + image)
		//		time.Sleep(time.Second * 1)
		tmp := filepath.Join(tempdir, fmt.Sprintf("%002d", i)) //fixme add extension
		err := ioutil.WriteFile(tmp, pngData, 0644)

		// compressImages
		if *compressBool {
			fmt.Println(tmp)
			img, err := png.Decode(bytes.NewReader(pngData))
			check(err)
			jpgData := compress(img)
			err = ioutil.WriteFile(tmp, jpgData, 0644)
			check(err)
		} else {
			err := ioutil.WriteFile(tmp, pngData, 0644)
			check(err)
		}

		// png not compressed

		check(err)
	}
	fmt.Println("Download complete")
	fmt.Println(tempdir)
	view(tempdir)
	return

}

// wird nicht benutzt
func mySpinner() {
	fmt.Printf("/")
	const t = 1
	i := 0
	for i < 10 {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(t * time.Second)
		}

	}
	/*
		s := spinner.New(spinner.CharSets[25], 300*time.Millisecond) // Build our new spinner

		s.Start()                   // Start the spinner
		time.Sleep(4 * time.Second) // Run for some time to simulate work
		s.Stop()
	*/
}

//
func makeSlice(imageString string) []string {
	imageString = strings.TrimLeft(imageString, "fileList = [")
	imageString = strings.TrimRight(imageString, "];")
	imageString = strings.Replace(imageString, "\"", "", -1)
	imageString = strings.Replace(imageString, " ", "", -1)
	images := strings.Split(imageString, ",")
	if limit < len(images) {
		return images[:limit]
	}
	return images[:len(images)]

}

// Baisc HTTP Auth
func authLoad(link string) string {
	username := usernameStr
	passwd := passwdStr
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("check connection")
	}
	req.SetBasicAuth(*username, *passwd)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		fmt.Println("check connection")
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		fmt.Println("check connection")
	}
	return string(bodyText)
}

func getImageListString(bodyString string) string {
	// Warschien ist response.Body bereits ein Reader
	s := bufio.NewScanner(strings.NewReader(bodyString))
	for s.Scan() {
		if strings.Contains(s.Text(), "fileList") {
			return s.Text()
		}
	}
	return ""

}

// compress replace all png images with compressed JPEGs
func compress(img image.Image) []byte {
	options := &jpeg.Options{Quality: COMPRESSION_LEVEL}
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, options)
	check(err)
	return buf.Bytes()
}

// Wenn irgendein Fehler auftauch -> Programm abstürzen lassen
// In der Regel haben höhere Go-Funktionen zwei(2) Rückgabewerte
// wert, err := packet.Funktion()
// check(err) //<- soll hier Fehler behandeln. Es geht besser

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Bildbetrachter
func view(dirstring string) {
	//Standard is eom eye of mate for me
	if *viewerStr == "none" {
		return
	}

	eom := exec.Command(*viewerStr, dirstring)
	eom.Run()
}

var Data []byte = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x20,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x73, 0x7a, 0x7a, 0xf4, 0x00, 0x00, 0x00,
	0x06, 0x62, 0x4b, 0x47, 0x44, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0xa0,
	0xbd, 0xa7, 0x93, 0x00, 0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00,
	0x00, 0x2e, 0x23, 0x00, 0x00, 0x2e, 0x23, 0x01, 0x78, 0xa5, 0x3f, 0x76,
	0x00, 0x00, 0x00, 0x07, 0x74, 0x49, 0x4d, 0x45, 0x07, 0xe2, 0x0b, 0x19,
	0x15, 0x00, 0x10, 0xca, 0xc3, 0x48, 0xd2, 0x00, 0x00, 0x00, 0x19, 0x74,
	0x45, 0x58, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x00, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x20, 0x77, 0x69, 0x74, 0x68, 0x20,
	0x47, 0x49, 0x4d, 0x50, 0x57, 0x81, 0x0e, 0x17, 0x00, 0x00, 0x01, 0x39,
	0x49, 0x44, 0x41, 0x54, 0x58, 0xc3, 0xed, 0x96, 0xb1, 0x4b, 0x42, 0x51,
	0x14, 0x87, 0xbf, 0x17, 0x81, 0x4e, 0x41, 0x09, 0xaf, 0x41, 0x5c, 0x2a,
	0x87, 0x37, 0xa5, 0xce, 0x05, 0x81, 0xd0, 0xe0, 0x10, 0x04, 0xbe, 0x31,
	0x2b, 0xa8, 0x86, 0x20, 0x70, 0xa9, 0xa1, 0x20, 0x1c, 0xa2, 0xb6, 0x96,
	0x88, 0x0a, 0x92, 0x5a, 0x83, 0x26, 0x09, 0x2c, 0x82, 0x96, 0xde, 0x12,
	0xb4, 0xf4, 0x0f, 0x38, 0x48, 0x06, 0x09, 0x36, 0x88, 0x18, 0x22, 0x74,
	0x1b, 0xce, 0x52, 0x4d, 0x49, 0xf9, 0x2e, 0xd1, 0x3d, 0x70, 0xb8, 0xf7,
	0x9e, 0xbb, 0x7c, 0xf0, 0x3b, 0xe7, 0x77, 0xaf, 0xd5, 0x8f, 0x52, 0x68,
	0x8c, 0x1e, 0x34, 0x87, 0x01, 0x30, 0x00, 0x7f, 0x07, 0x20, 0x9a, 0x80,
	0x17, 0xf5, 0x39, 0x6b, 0x6f, 0xf0, 0xf4, 0x0a, 0x0f, 0x65, 0x38, 0x3a,
	0x83, 0x3e, 0xbb, 0x73, 0x80, 0xde, 0x9f, 0xd0, 0x5b, 0x16, 0x04, 0x83,
	0x10, 0x89, 0x48, 0x8e, 0x44, 0x21, 0x99, 0xe8, 0xb2, 0x04, 0xad, 0x16,
	0x0c, 0x58, 0x92, 0xce, 0x30, 0x2c, 0x2d, 0x80, 0xe7, 0xc9, 0x5d, 0x3c,
	0x0e, 0x99, 0x15, 0x1f, 0x7b, 0xe0, 0xb9, 0x04, 0xe7, 0x79, 0x98, 0x1a,
	0x87, 0x6a, 0x55, 0x6a, 0xa3, 0x31, 0xcd, 0x4d, 0xd8, 0x6c, 0xfa, 0xd8,
	0x03, 0x83, 0x43, 0x30, 0x36, 0x01, 0xb3, 0xf3, 0x60, 0xdb, 0x22, 0xcf,
	0x45, 0xa1, 0xcb, 0x00, 0x81, 0x80, 0x4c, 0xc0, 0xd7, 0x68, 0x34, 0xe0,
	0xf0, 0x00, 0xee, 0xae, 0x35, 0x48, 0xa0, 0x14, 0x14, 0x8b, 0xb0, 0xbd,
	0xe6, 0x83, 0x11, 0x7d, 0x9c, 0x82, 0xc9, 0x24, 0xe4, 0x8f, 0xa5, 0xe6,
	0xba, 0xb0, 0xb9, 0xeb, 0xb3, 0x13, 0xde, 0xdf, 0xc0, 0xea, 0x22, 0x9c,
	0x9e, 0xc8, 0x79, 0x26, 0xd3, 0xb9, 0x19, 0xfd, 0x8a, 0x04, 0xeb, 0xcb,
	0x50, 0xa9, 0x40, 0x28, 0x04, 0x1b, 0x5b, 0x9a, 0xc6, 0xf0, 0xea, 0x52,
	0xd6, 0x54, 0x4a, 0x13, 0xc0, 0x4e, 0x0e, 0xea, 0x75, 0x08, 0x87, 0x21,
	0x9b, 0xd3, 0x00, 0x50, 0x7b, 0x04, 0xef, 0x56, 0xf6, 0x69, 0x57, 0x93,
	0x13, 0xee, 0xef, 0x41, 0xbb, 0x0d, 0x8e, 0x03, 0xd3, 0x73, 0xdf, 0x7c,
	0xd0, 0xcc, 0xaf, 0xd8, 0x00, 0x18, 0x00, 0x03, 0xf0, 0xef, 0x01, 0xde,
	0x01, 0x63, 0x15, 0x50, 0x36, 0xf5, 0x77, 0x56, 0x7c, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}
