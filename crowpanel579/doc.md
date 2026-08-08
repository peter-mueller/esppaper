# 5.79 Panel

By https://github.com/samperk1/esphome-crowpanel-579

The 5.79" panel uses two SSD1683 driver chips wired to the same SPI bus — one chip drives the left half, the other drives the right half. Both chips share CS, DC, RST, and BUSY lines. They are differentiated by their command sets:

Slave (physical right half, columns 0–399): commands 0x91, 0xA4, 0xA6, 0xC4/C5/CE/CF
Master (physical left half, columns 392–791): commands 0x11, 0x24, 0x26, 0x44/45/4E/4F
The 8-pixel overlap at columns 392–399 (byte 49 in the buffer) is shared between both chips and provides seam alignment.

The single framebuffer is 99 bytes × 272 rows = 26,928 bytes. On each display() call, bytes 0–49 per row go to the slave and bytes 49–98 per row (reversed) go to the master. A full refresh takes approximately 3 seconds.