RESULTS_START  EQU $c000
RESULTS_N_ROWS EQU 6

include "base.inc"

; Channel 2 is aligned to the APU's enable time, not the CPU's start
; time. Since we're running in a 1MHz world now, further tests will
; use single speed mode.


CorrectResults:

db $00, $00, $00, $00, $00, $00, $00, $00
db $00, $00, $00, $00, $00, $80, $80, $80
db $00, $00, $00, $00, $00, $00, $00, $00
db $00, $00, $00, $00, $00, $80, $80, $80
db $00, $00, $00, $00, $00, $00, $00, $00
db $00, $00, $00, $00, $00, $80, $80, $80

SubTest: MACRO
    xor a
    ldh [rNR52], a
    cpl
    nops \2
    ldh [rNR52], a
    
    ld hl, rPCM12
    ldh [rNR23], a
    ld a, $80
    ldh [rNR21], a
    ldh [rNR22], a
    ld a, $87
    ldh [rNR24], a
    
    nops \1
    
    ld a, [hl]
    call StoreResult
    ENDM

RunTest:
    ld a, 1
    ldh [rKEY1], a
    stop
    
    ld de, $c000
    
    SubTest $0, 0 
    SubTest $1, 0
    SubTest $2, 0
    SubTest $3, 0
    SubTest $4, 0
    SubTest $5, 0
    SubTest $6, 0
    SubTest $7, 0
    
    SubTest $8, 0
    SubTest $9, 0
    SubTest $a, 0
    SubTest $b, 0
    SubTest $c, 0
    SubTest $d, 0
    SubTest $e, 0
    SubTest $f, 0
    
    SubTest $0, 1
    SubTest $1, 1
    SubTest $2, 1
    SubTest $3, 1
    SubTest $4, 1
    SubTest $5, 1
    SubTest $6, 1
    SubTest $7, 1
    
    SubTest $8, 1
    SubTest $9, 1
    SubTest $a, 1
    SubTest $b, 1
    SubTest $c, 1
    SubTest $d, 1
    SubTest $e, 1
    SubTest $f, 1
    
    SubTest $0, 3 
    SubTest $1, 3
    SubTest $2, 3
    SubTest $3, 3
    SubTest $4, 3
    SubTest $5, 3
    SubTest $6, 3
    SubTest $7, 3
    
    SubTest $8, 3
    SubTest $9, 3
    SubTest $a, 3
    SubTest $b, 3
    SubTest $c, 3
    SubTest $d, 3
    SubTest $e, 3
    SubTest $f, 3
        
    ret
    
    
StoreResult::
    ld [de], a
    inc de
    ret
    
    CGB_MODE
