# Pitch Buzzer / Ticker

## Raspberry Pi

### root-Filesystem
First of all enlarge the root filesystem with raspi-config, to use the whole size of the SD card.

### Packages

Buzzer:

Raspbian

    apt-get install mercurial \
        libgles1-mesa-dev \
        libgles2-mesa-dev \
        libperl-dev \
        libgtk2.0-dev

RHEL:

    yum install pango.x86_64 \
        pango-devel.x86_64 \ 
        gdk-pixbuf2.x86_64 \
        gdk-pixbuf2-devel.x86_64 \
        gtk2-devel.x86_64 


Ticker:

### Console
For the Ticker you will need the /dev/ttyAMA0, which is in use for the serial console. Remove appropriate console entry in /boot/cmdline.txt and reboot.

change:

    dwc_otg.lpm_enable=0 console=ttyAMA0,115200 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait

to:

    dwc_otg.lpm_enable=0 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait

see also: http://0pointer.de/blog/projects/serial-console.html

### WLAN

/etc/network/interfaces

    allow-hotplug wlan0
    iface wlan0 inet dhcp
        wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf

/etc/wpa_supplicant/wpa_supplicant.conf

    network={
      ssid="INSERT"
      psk="INSERT"
      proto=RSN
      key_mgmt=WPA-PSK
      pairwise=CCMP
      auth_alg=OPEN
    }

### Golang
Install the latest go version from:
* http://dave.cheney.net/unofficial-arm-tarballs

    tar -C /usr/local -xzf go1.5.2.linux-arm.tar.gz

$HOME/.profile

    if [ -d /usr/local/go/bin ]; then
        PATH="/usr/local/go/bin:$PATH"
    fi

    export PATH
    export GOPATH=$HOME/golang
    export CDPATH=$HOME/golang/src/gitlab.com/marcsauter

#### Packages

Buzzer:
* github.com/luismesas/goPi/piface
* github.com/luismesas/goPi/spi
* github.com/mattn/go-gtk

Ticker:
* github.com/tarm/serial

### PiFace Digital 2
* https://godoc.org/github.com/luismesas/goPi/piface
* http://www.piface.org.uk/assets/docs/PiFace-Digital2_getting-started.pdf

### Ticker
Pin Assignment:

    Ticker Plug RJ11/RS-485 (front view)
                                               ----------
    RX  / connect to GPIO 15 (RXD)   1 |-       |
    TX  / connect to GPIO 14 (TXD)   2 |-       ---
    DTR / not connected              3 |-         |
    GND / connect to GND             4 |-         |
    RTS / not connected              5 |-       ---
    not used                         6 |-       |
                                               ----------

    Raspi Plug GPIO (top view):
       o     o     o     o    o-- 3.3V
    -------------------
    | RXD | TXD | GND |  o    o-- 5V
    -------------------

## Web service
buzzer-ws on Google Appengine
