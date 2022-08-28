## GAMPBOY EMULATOR

That's my first project in Golang, so please be kind to me
Probably there are a lot of typos around the code and so many optimizations
could be done , but don't care for now :3

Probably one of the most satisfying and interesting project I have ever done

What have I learned?

- Go hates so much structs and map[string]

#### MBC supported

- [x] `MBC1`
- [x] `MBC3`

#### Blargg's tests

- [x] `cpu_instrs`
- [x] `instr_timing`
- [x] `mem-timing` and `mem-timing-2`
- [ ] `oam_bug/1-lcd_sync`
  - that's pretty weird, the whole test is based on two delay  
     `delay 109`  
     `delay 110`  
    if you see the code that those asm instructions generate look like  
     `LD A,0x50 ; delay 109`  
     `CALL 0xC120`  
     ....  
     and  
     `LD A,0x51 ; delay 110`  
     `CALL 0xC120`  
     ...  
     then the procedure will `SUB 0x5` and loop until `JR NC,0xFC`  
     but actually both 0x50 and 0x51 will do the same number of loops  
     so ... plz someone tell me how to understand this test  
     clearly if I add +6 to the second delay, this would be enough to make  
     the test passed, because the loop will do one more cycle
- [ ] `halt_bug`

  - correct output

    ![this](https://felixweichselgartner.github.io/assets/img/5cd5efaf-c9b8-4b57-b3fa-1b48d2111440.png)

#### Mooneye's tests

- [x] `interrupt/ie_push`
- [x] `acceptance/ie_timing`
- [x] `acceptance/rapid_di_ei`
- [ ] `acceptance/div_timing`
- [x] `acceptance/oam_dma_start`
- [x] `acceptance/oam_dma_restart`
- [x] `acceptance/oam_dma_timing`
- [x] `acceptance/push_timing`
- [ ] `acceptance/pop_timing`
- [x] `oam_dma/basic`
- [x] `oam_dma/reg_read`
- [x] `oam_dma/sources-GS`
- [x] `ppu/stat_lyc_onoff`

#### Mattcurrie's tests

- [x] `dmg_acid2`
- [x] `cgb_acid2`

#### SameSuite's tests

- [x] `ppu/blocking_bgpi_increase`
- [x] `dma/gdma_addr_mask`
- [x] `dma/gbc_dma_cont`
- [x] `dma/hdma_lcd_off`

#### Thanks to

- http://imrannazar.com/GameBoy-Emulation-in-JavaScript
- https://github.com/rockytriton/LLD_gbemu
- https://github.com/retrio/gb-test-roms
- https://gbdev.io/
- https://ia903208.us.archive.org/9/items/GameBoyProgManVer1.1/GameBoyProgManVer1.1.pdf
- https://blog.tigris.fr/2019/09/15/writing-an-emulator-the-first-pixel/
- https://gekkio.fi/files/gb-docs/gbctr.pdf
- https://www.youtube.com/watch?v=HyzD8pNlpwI
- https://github.com/AntonioND/giibiiadvance/tree/master/docs
