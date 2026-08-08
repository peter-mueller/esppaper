package crowpanel579

import (
	"image/color"
	"log/slog"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

const (
	Height  = 272
	Width   = 792
	HalfRow = 50
	FullRow = 99 // 1 Byte Overlap
)

const (
	SWReset               = 0x12
	SenseTemperature      = 0x18
	LoadWaveform          = 0x22
	MasterActivation      = 0x20 // blink refresh
	TemperatureRegister   = 0x1A
	BorderWaveformControl = 0x3C
	MasterNewRAM          = 0x24
	MasterOldRAM          = 0x26
	SlaveNewRAM           = 0xA4
	SlaveOldRAM           = 0xA6
)

type CrowPanel579 struct {
	bus drivers.SPI

	cs   machine.Pin // Chip Select
	dc   machine.Pin // Data/Command Control
	rst  machine.Pin // Hardware Reset
	busy machine.Pin // Busy

	buffer [FullRow * Height]byte
}

func (c *CrowPanel579) xyIndex(x, y int16) (byteIndex uint32, bitIndex uint8) {
	byteIndex = (uint32(y)*Width + uint32(x)) / 8
	bitIndex = uint8(x % 8)
	return
}

func NewCrowPanel579(bus drivers.SPI, csPin, dcPin, rstPin, busyPin machine.Pin) CrowPanel579 {

	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busyPin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	return CrowPanel579{
		bus:    bus,
		cs:     csPin,
		dc:     dcPin,
		rst:    rstPin,
		busy:   busyPin,
		buffer: [26928]byte{},
	}
}

var white = color.Gray{0xFF}

func (c *CrowPanel579) SetPixel(x int16, y int16, col color.Gray) {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return
	}
	byteIndex, bitIndex := c.xyIndex(x, y)

	if col == white {
		c.buffer[byteIndex] |= 0x80 >> bitIndex
	} else {
		c.buffer[byteIndex] &^= 0x80 >> bitIndex
	}
}

func (c *CrowPanel579) Configure() {
	slog.Info("1. Power On")
	time.Sleep(10 * time.Millisecond)
	slog.Info("2. Set Initial Configuration")
	c.HWReset()
	c.WaitBusy()
	c.SendCommand(SWReset)
	c.WaitBusy()

	slog.Info("3. Send Initialization Code")
	c.SendCommand(BorderWaveformControl)
	c.SendData(0x01) // border waveform

	slog.Info("4. Load Waveform LUT")
	c.SendCommand(SenseTemperature)
	c.SendData(0x80) // use internal temperature sensor
	c.SendCommand(LoadWaveform)
	c.SendData(0xB1) // enable clock, CP, load temp
	c.SendCommand(MasterActivation)
	c.WaitBusy()
	c.SendCommand(TemperatureRegister)
	c.SendData(0x64)
	c.SendData(0x00)

	c.SendCommand(LoadWaveform)
	c.SendData(0x91) // Load temperature value
	c.SendCommand(MasterActivation)
	c.WaitBusy()

	c.SendCommand(BorderWaveformControl) // Set panel border
	c.SendData(0x3)
	c.WaitBusy()

	slog.Info("5. Write Image and Drive Display Panel")
	// Clear both RAM banks to white on both chips.
	// Old RAM must match New RAM or the anti-ghosting waveform produces static.
	c.SetRamMaster()
	c.SendCommand(MasterNewRAM)
	c.WriteRamHalf(0x00) // master new RAM = white
	c.SetRamMaster()
	c.SendCommand(MasterOldRAM)
	c.WriteRamHalf(0x00) // master old RAM = white
	c.SetRamSlave()
	c.SendCommand(SlaveNewRAM)
	c.WriteRamHalf(0x00) // slave new RAM = white
	c.SetRamSlave()
	c.SendCommand(SlaveOldRAM)
	c.WriteRamHalf(0x00) // slave old RAM = white

	// Full refresh to flush panel to known-white state
	c.SendCommand(0x22)
	c.SendData(0xF7)
	c.SendCommand(0x20)
	c.WaitBusy()
}

// SendCommand sends a command to the display
func (d *CrowPanel579) SendCommand(command uint8) {
	d.sendDataCommand(true, command)
}

// SendData sends a data byte to the display
func (d *CrowPanel579) SendData(data uint8) {
	d.sendDataCommand(false, data)
}

func (c *CrowPanel579) WriteRamHalf(data uint8) {
	c.dc.High()
	c.cs.Low()
	for i := 0; i < HalfRow*8*Height; i++ {
		c.bus.Transfer(data)
	}
	c.cs.High()
}

// WaitUntilIdle waits until the display is ready
func (d *CrowPanel579) WaitBusy() {
	for d.busy.Get() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *CrowPanel579) Display() {
	slog.Info("Display RAM SLAVE")
	c.SetRamSlave()
	c.SendCommand(SlaveNewRAM)
	c.dc.High()
	c.cs.Low()
	for y := 0; y < Height; y++ {
		rowStart := y * FullRow
		for b := 0; b < HalfRow; b++ {
			data := c.buffer[rowStart+b]
			c.bus.Transfer(data)
		}
	}
	c.cs.High()

	slog.Info("Display RAM Master")
	c.SetRamMaster()
	c.SendCommand(MasterNewRAM)
	c.dc.High()
	c.cs.Low()
	for y := 0; y < Height; y++ {
		rowStart := y * FullRow
		for b := 49; b < FullRow; b++ {
			data := c.buffer[rowStart+b]
			c.bus.Transfer(data)
		}
	}
	c.cs.High()

	slog.Info("Fast Update")
	c.SendCommand(0x22)
	c.SendData(0xF7)
	c.SendCommand(MasterActivation)
	c.WaitBusy()
}

func (c *CrowPanel579) SetRamMaster() {
	c.SendCommand(0x11) // Data Entry mode setting
	c.SendData(0x02)
	c.SendCommand(0x44)
	c.SendData(0x31)    //400/8-1
	c.SendData(0x00)    // XStart, POR = 00h
	c.SendCommand(0x45) // Set Ram Y- address  Start / End position
	c.SendData(0x00)
	c.SendData(0x00) // YEnd L
	c.SendData(0x0F)
	c.SendData(0x01) //300-1

	c.SendCommand(0x4E) //Set RAM X address counter
	c.SendData(0x31)
	c.SendCommand(0x4F)
	c.SendData(0x00)
	c.SendData(0x00)

}

func (c *CrowPanel579) SetRamSlave() {
	c.SendCommand(0x91) // SlaveDateEntryMode
	c.SendData(0x03)
	c.SendCommand(0xC4)
	c.SendData(0x00)
	c.SendData(0x31)
	c.SendCommand(0xC5)
	c.SendData(0x00)
	c.SendData(0x00)
	c.SendData(0x0F)
	c.SendData(0x01)

	c.SendCommand(0xCE) // SlaveDateEntryMode
	c.SendData(0x00)
	c.SendCommand(0xCF)
	c.SendData(0x00)
	c.SendData(0x00)
}

// sendDataCommand sends image data or a command to the screen
func (d *CrowPanel579) sendDataCommand(isCommand bool, data uint8) {
	if isCommand {
		d.dc.Low()
	} else {
		d.dc.High()
	}
	d.cs.Low()
	d.bus.Transfer(data)
	d.cs.High()
}

func (c *CrowPanel579) HWReset() {
	time.Sleep(10 * time.Millisecond)
	c.rst.Low()
	time.Sleep(10 * time.Millisecond)
	c.rst.High()
	time.Sleep(10 * time.Millisecond)
	c.WaitBusy()
}
