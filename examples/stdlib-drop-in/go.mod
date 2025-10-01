module stdlib-drop-in

go 1.23

require github.com/agilira/flash-flags/stdlib v0.0.0

require github.com/agilira/flash-flags v1.0.4 // indirect

replace github.com/agilira/flash-flags/stdlib => ../../stdlib

replace github.com/agilira/flash-flags => ../..
