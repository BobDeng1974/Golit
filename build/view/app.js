function tasmota_on(feed) {
    fetch('mqtt/cmd/' + feed + '/power/on').then(function (response) {
        let stat = document.getElementById("status_" + feed);
        stat.classList.remove("status_0");
        stat.classList.add("status_1");
    });
}

function tasmota_off(feed) {
    fetch('mqtt/cmd/' + feed + '/power/off').then(function (response) {
        let stat = document.getElementById("status_" + feed);
        stat.classList.remove("status_1");
        stat.classList.add("status_0");
    });
}

function sys_on() {
    fetch('on').then(function (response) {
        app_reload()
    });
}

function sys_off() {
    fetch('off').then(function (response) {
        app_reload()
    });
}

function hue_pair() {
    fetch('hue/pair').then(function (response) {
        location.href = "/"
    });
}

function hue_on(uid) {
    fetch('hue/light/'+uid+'/on').then(function (response) {
        let stat = document.getElementById("status_" + uid);
        stat.classList.remove("status_false");
        stat.classList.add("status_true");
    });
}
function hue_off(uid) {
    fetch('hue/light/'+uid+'/off').then(function (response) {
        let stat = document.getElementById("status_" + uid);
        stat.classList.remove("status_true");
        stat.classList.add("status_false");
    });
}

function hue_setScene() {
    let scene = document.getElementById("scenes").value;
    fetch('hue/scene/'+scene).then(function (response) {
    });
}

function hue_apply(uid) {
    let ct = document.getElementById(uid + "_ct").value;
    let bri = document.getElementById(uid + "_bri").value;
    post_data('hue/light/'+uid+'/update', {Ct: parseInt(ct, 10), Bri: parseInt(bri,10)}).then(data => {
        console.log(data.result);
        window.location.href = "/";
    }).catch(error => console.error(error));
}

function tasmota_setColor(feed) {
    let color = document.getElementById(feed + "_color").value;
    let white = document.getElementById(feed + "_white").value;
    let whiteVal = parseInt(white,10);
    if(whiteVal < 15) {
        white = "0" + parseInt(white,10).toString(16)
    }
    else {
        white = parseInt(white,10).toString(16)
    }
    console.log(white)
    color = color.replace("#", "*");
    fetch('mqtt/cmd/' + feed + '/Color/' + color+white).then(function (response) {});
}


function set_slider_hex(feed, hex) {
    document.getElementById(feed + "_white").value = parseInt(hex,16).toString();
}

function set_input_from_cie(uid, x, y, bright) {
    document.getElementById(uid + "_color").value = cie_to_hex(x,y,bright);
}

function tasmota_add() {
    let feed = document.getElementById("new_feed").value;
    post_data('add/tasmota', {
        Feed: feed
    }).then(data => {
        console.log(data.result);
        window.location.href = "/";
    }).catch(error => console.error(error));
}

function app_reload() {
    location.reload();
}

function tasmota_delete(feed) {
    post_data('del/tasmota', {
        Feed: feed
    }).then(data => {
        console.log(data.result);
        window.location.href = "/";
    }).catch(error => console.error(error));
}

function app_cancel() {
    window.location.href = "/";
}

function post_data(url = '', data = {}) {
    // Default options are marked with *
    return fetch(url, {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, cors, *same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            'Content-Type': 'application/json',
            // 'Content-Type': 'application/x-www-form-urlencoded',
        },
        redirect: 'follow', // manual, *follow, error
        referrer: 'no-referrer', // no-referrer, *client
        body: JSON.stringify(data), // body data type must match "Content-Type" header
    })
        .then(response => response.json()); // parses JSON response into native JavaScript objects
}

function tab_nav(target, uid) {
    let dest = document.getElementById(target + "_" + uid);
    let dest_tab = document.getElementById(dest.id + "_tab");
    let active_tab = document.getElementById("tabs" + "_" + uid).querySelector(".is-active");
    let active = document.getElementById(active_tab.id.replace("_tab",""));

    active.hidden = true;
    dest.hidden = false;
    active_tab.classList.remove("is-active");
    dest_tab.classList.add("is-active");
}

/*
With these functions you can convert the CIE color space to the RGB color space and vice versa.

The developer documentation for Philips Hue provides the formulas used in the code below:
https://developers.meethue.com/documentation/color-conversions-rgb-xy

I've used the formulas and Objective-C example code and transfered it to JavaScript.


Examples:

var rgb = cie_to_rgb(0.6611, 0.2936)
var cie = rgb_to_cie(255, 39, 60)

------------------------------------------------------------------------------------

The MIT License (MIT)

Copyright (c) 2017 www.usolved.net
Published under https://github.com/usolved/cie-rgb-converter

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/


function cie_to_hex(x, y, brightness) {
    let rgb = cie_to_rgb(x,y,brightness)
    return rgb[0].toString() +","+ rgb[1].toString() + "," + rgb[2].toString()
}

function wrt(v) {
    document.write(v)
}

/**
 * Converts CIE color space to RGB color space
 * @param {Number} x
 * @param {Number} y
 * @param {Number} brightness - Ranges from 1 to 254
 * @return {Array} Array that contains the color values for red, green and blue
 */
function cie_to_rgb(x, y, brightness)
{
    //Set to maximum brightness if no custom value was given (Not the slick ECMAScript 6 way for compatibility reasons)
    if (brightness === undefined) {
        brightness = 254;
    }

    var z = 1.0 - x - y;
    var Y = (brightness / 254).toFixed(2);
    var X = (Y / y) * x;
    var Z = (Y / y) * z;

    //Convert to RGB using Wide RGB D65 conversion
    var red 	=  X * 1.656492 - Y * 0.354851 - Z * 0.255038;
    var green 	= -X * 0.707196 + Y * 1.655397 + Z * 0.036152;
    var blue 	=  X * 0.051713 - Y * 0.121364 + Z * 1.011530;

    //If red, green or blue is larger than 1.0 set it back to the maximum of 1.0
    if (red > blue && red > green && red > 1.0) {

        green = green / red;
        blue = blue / red;
        red = 1.0;
    }
    else if (green > blue && green > red && green > 1.0) {

        red = red / green;
        blue = blue / green;
        green = 1.0;
    }
    else if (blue > red && blue > green && blue > 1.0) {

        red = red / blue;
        green = green / blue;
        blue = 1.0;
    }

    //Reverse gamma correction
    red 	= red <= 0.0031308 ? 12.92 * red : (1.0 + 0.055) * Math.pow(red, (1.0 / 2.4)) - 0.055;
    green 	= green <= 0.0031308 ? 12.92 * green : (1.0 + 0.055) * Math.pow(green, (1.0 / 2.4)) - 0.055;
    blue 	= blue <= 0.0031308 ? 12.92 * blue : (1.0 + 0.055) * Math.pow(blue, (1.0 / 2.4)) - 0.055;


    //Convert normalized decimal to decimal
    red 	= Math.round(red * 255);
    green 	= Math.round(green * 255);
    blue 	= Math.round(blue * 255);

    if (isNaN(red))
        red = 0;

    if (isNaN(green))
        green = 0;

    if (isNaN(blue))
        blue = 0;


    return [red, green, blue];
}


/**
 * Converts RGB color space to CIE color space
 * @param {Number} red
 * @param {Number} green
 * @param {Number} blue
 * @return {Array} Array that contains the CIE color values for x and y
 */
function rgb_to_cie(red, green, blue)
{
    //Apply a gamma correction to the RGB values, which makes the color more vivid and more the like the color displayed on the screen of your device
    var red 	= (red > 0.04045) ? Math.pow((red + 0.055) / (1.0 + 0.055), 2.4) : (red / 12.92);
    var green 	= (green > 0.04045) ? Math.pow((green + 0.055) / (1.0 + 0.055), 2.4) : (green / 12.92);
    var blue 	= (blue > 0.04045) ? Math.pow((blue + 0.055) / (1.0 + 0.055), 2.4) : (blue / 12.92);

    //RGB values to XYZ using the Wide RGB D65 conversion formula
    var X 		= red * 0.664511 + green * 0.154324 + blue * 0.162028;
    var Y 		= red * 0.283881 + green * 0.668433 + blue * 0.047685;
    var Z 		= red * 0.000088 + green * 0.072310 + blue * 0.986039;

    //Calculate the xy values from the XYZ values
    var x 		= (X / (X + Y + Z)).toFixed(4);
    var y 		= (Y / (X + Y + Z)).toFixed(4);

    if (isNaN(x))
        x = 0;

    if (isNaN(y))
        y = 0;


    return [x, y];
}
