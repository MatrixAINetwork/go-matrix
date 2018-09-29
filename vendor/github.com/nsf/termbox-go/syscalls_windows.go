// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs -- -DUNICODE syscalls.go

package termbox

const (
	foreground_blue          = 0x1
	foreground_green         = 0x2
	foreground_red           = 0x4
	foreground_intensity     = 0x8
	background_blue          = 0x10
	background_green         = 0x20
	background_red           = 0x40
	background_intensity     = 0x80
	std_input_handle         = -0xa
	std_output_handle        = -0xb
	key_event                = 0x1
	mouse_event              = 0x2
	window_buffer_size_event = 0x4
	enable_window_input      = 0x8
	enable_mouse_input       = 0x10
	enable_extended_flags    = 0x80

	vk_f1          = 0x70
	vk_f2          = 0x71
	vk_f3          = 0x72
	vk_f4          = 0x73
	vk_f5          = 0x74
	vk_f6          = 0x75
	vk_f7          = 0x76
	vk_f8          = 0x77
	vk_f9          = 0x78
	vk_f10         = 0x79
	vk_f11         = 0x7a
	vk_f12         = 0x7b
	vk_insert      = 0x2d
	vk_delete      = 0x2e
	vk_home        = 0x24
	vk_end         = 0x23
	vk_pgup        = 0x21
	vk_pgdn        = 0x22
	vk_arrow_up    = 0x26
	vk_arrow_down  = 0x28
	vk_arrow_left  = 0x25
	vk_arrow_right = 0x27
	vk_backspace   = 0x8
	vk_tab         = 0x9
	vk_enter       = 0xd
	vk_esc         = 0x1b
	vk_space       = 0x20

	left_alt_pressed   = 0x2
	left_ctrl_pressed  = 0x8
	right_alt_pressed  = 0x1
	right_ctrl_pressed = 0x4
	shift_pressed      = 0x10

	generic_read            = 0x80000000
	generic_write           = 0x40000000
	console_textmode_buffer = 0x1
)
