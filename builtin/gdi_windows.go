package builtin

import (
	"unsafe"

	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/verify"
	"golang.org/x/sys/windows"
)

var gdi32 = windows.NewLazyDLL("gdi32.dll")

const LF_FACESIZE = 32

type LOGFONTA struct {
	lfHeight         int32
	lfWidth          int32
	lfEscapement     int32
	lfOrientation    int32
	lfWeight         int32
	lfItalic         byte
	lfUnderline      byte
	lfStrikeOut      byte
	lfCharSet        byte
	lfOutPrecision   byte
	lfClipPrecision  byte
	lfQuality        byte
	lfPitchAndFamily byte
	lfFaceName       [LF_FACESIZE]byte
}

type TEXTMETRIC struct {
	Height           int32
	Ascent           int32
	Descent          int32
	InternalLeading  int32
	ExternalLeading  int32
	AveCharWidth     int32
	MaxCharWidth     int32
	Weight           int32
	Italic           byte
	Underlined       byte
	StruckOut        byte
	FirstChar        byte
	LastChar         byte
	DefaultChar      byte
	BreakChar        byte
	PitchAndFamily   byte
	CharSet          byte
	Overhang         int32
	DigitizedAspectX int32
	DigitizedAspectY int32
}

// dll Gdi32:CreateFontIndirect(LOGFONT* lf) gdiobj
// e.g. CreateFontIndirect(#(lfFaceName: "Segoe UI", lfHeight: -27))
var createFontIndirect = gdi32.NewProc("CreateFontIndirectA")
var _ = builtin1("CreateFontIndirect(logfont)", func(a Value) Value {
	f := LOGFONTA{
		lfHeight:         getInt32(a, "lfHeight"),
		lfWidth:          getInt32(a, "lfWidth"),
		lfEscapement:     getInt32(a, "lfEscapement"),
		lfOrientation:    getInt32(a, "lfOrientation"),
		lfWeight:         getInt32(a, "lfWeight"),
		lfItalic:         byte(getInt(a, "lfItalic")),
		lfUnderline:      byte(getInt(a, "lfUnderline")),
		lfStrikeOut:      byte(getInt(a, "lfStrikeOut")),
		lfCharSet:        byte(getInt(a, "lfCharSet")),
		lfOutPrecision:   byte(getInt(a, "lfOutPrecision")),
		lfClipPrecision:  byte(getInt(a, "lfClipPrecision")),
		lfQuality:        byte(getInt(a, "lfQuality")),
		lfPitchAndFamily: byte(getInt(a, "lfPitchAndFamily")),
	}
	copy(f.lfFaceName[:], ToStr(a.Get(nil, SuStr("lfFaceName"))))
	rtn, _, _ := createFontIndirect.Call(uintptr(unsafe.Pointer(&f)))
	return intRet(rtn)
})

// dll bool Gdi32:GetTextMetrics(pointer hdc, TEXTMETRIC* tm)
var getTextMetrics = gdi32.NewProc("GetTextMetricsA")
var _ = builtin2("GetTextMetrics(hdc, tm)",
	func(a, b Value) Value {
		var tm TEXTMETRIC
		rtn, _, _ := getTextMetrics.Call(
			intArg(a),
			uintptr(unsafe.Pointer(&tm)))
		b.Put(nil, SuStr("Height"), IntVal(int(tm.Height)))
		b.Put(nil, SuStr("Ascent"), IntVal(int(tm.Ascent)))
		b.Put(nil, SuStr("Descent"), IntVal(int(tm.Descent)))
		b.Put(nil, SuStr("InternalLeading"), IntVal(int(tm.InternalLeading)))
		b.Put(nil, SuStr("ExternalLeading"), IntVal(int(tm.ExternalLeading)))
		b.Put(nil, SuStr("AveCharWidth"), IntVal(int(tm.AveCharWidth)))
		b.Put(nil, SuStr("MaxCharWidth"), IntVal(int(tm.MaxCharWidth)))
		b.Put(nil, SuStr("Italic"), IntVal(int(tm.Italic)))
		b.Put(nil, SuStr("Underlined"), IntVal(int(tm.Underlined)))
		b.Put(nil, SuStr("StruckOut"), IntVal(int(tm.StruckOut)))
		b.Put(nil, SuStr("FirstChar"), IntVal(int(tm.FirstChar)))
		b.Put(nil, SuStr("LastChar"), IntVal(int(tm.LastChar)))
		b.Put(nil, SuStr("DefaultChar"), IntVal(int(tm.DefaultChar)))
		b.Put(nil, SuStr("BreakChar"), IntVal(int(tm.BreakChar)))
		b.Put(nil, SuStr("PitchAndFamily"), IntVal(int(tm.PitchAndFamily)))
		b.Put(nil, SuStr("CharSet"), IntVal(int(tm.CharSet)))
		b.Put(nil, SuStr("Overhang"), IntVal(int(tm.Overhang)))
		b.Put(nil, SuStr("DigitizedAspectX"), IntVal(int(tm.DigitizedAspectX)))
		b.Put(nil, SuStr("DigitizedAspectY"), IntVal(int(tm.DigitizedAspectY)))
		return boolRet(rtn)
	})

var getStockObject = gdi32.NewProc("GetStockObject")
var _ = builtin1("GetStockObject(i)", func(a Value) Value {
	rtn, _, _ := getStockObject.Call(intArg(a))
	return intRet(rtn)
})

var getDeviceCaps = gdi32.NewProc("GetDeviceCaps")
var _ = builtin2("GetDeviceCaps(hdc, nIndex)", func(a, b Value) Value {
	rtn, _, _ := getDeviceCaps.Call(
		intArg(a),
		intArg(b))
	return intRet(rtn)
})

// dll GDI32:BitBlt(pointer hdcDest, long nXDest, long nYDest, long nWidth, long nHeight, pointer hdcSrc, long nXSrc, long nYSrc, long dwRop) bool
var bitBlt = gdi32.NewProc("BitBlt")
var _ = builtin("BitBlt(hdc, x, y, cx, cy, hdcSrc, x1, y1, rop)",
	func(_ *Thread, a []Value) Value {
		rtn, _, _ := bitBlt.Call(
			intArg(a[0]),
			intArg(a[1]),
			intArg(a[2]),
			intArg(a[3]),
			intArg(a[4]),
			intArg(a[5]),
			intArg(a[6]),
			intArg(a[7]),
			intArg(a[8]))
		return intRet(rtn)
	})

// dll Gdi32:CreateCompatibleBitmap(pointer hdc, long nWidth, long nHeight) gdiobj
var createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
var _ = builtin3("CreateCompatibleBitmap(hdc, cx, cy)", func(a, b, c Value) Value {
	rtn, _, _ := createCompatibleBitmap.Call(
		intArg(a),
		intArg(b),
		intArg(c))
	return intRet(rtn)
})

// dll Gdi32:CreateCompatibleDC(pointer hdc) pointer
var createCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
var _ = builtin1("CreateCompatibleDC(hdc)", func(a Value) Value {
	rtn, _, _ := createCompatibleDC.Call(intArg(a))
	return intRet(rtn)
})

// dll Gdi32:CreateSolidBrush(long rgb) gdiobj
var createSolidBrush = gdi32.NewProc("CreateSolidBrush")
var _ = builtin1("CreateSolidBrush(i)", func(a Value) Value {
	rtn, _, _ := createSolidBrush.Call(intArg(a))
	return intRet(rtn)
})

// dll pointer Gdi32:SelectObject(pointer hdc, pointer obj)
var selectObject = gdi32.NewProc("SelectObject")
var _ = builtin2("SelectObject(hdc, obj)",
	func(a, b Value) Value {
		rtn, _, _ := selectObject.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll bool Gdi32:GetTextExtentPoint32(pointer hdc, [in] string text, long len, POINT* p)
var getTextExtentPoint32 = gdi32.NewProc("GetTextExtentPoint32A")
var _ = builtin4("GetTextExtentPoint32(hdc, text, len, p)",
	func(a, b, c, d Value) Value {
		var pt POINT
		rtn, _, _ := getTextExtentPoint32.Call(
			intArg(a),
			uintptr(stringArg(b)),
			uintptr(len(ToStr(b))),
			uintptr(unsafe.Pointer(&pt)))
		pointToOb(&pt, d)
		return boolRet(rtn)
	})

var getTextFace = gdi32.NewProc("GetTextFaceA")
var _ = builtin1("GetTextFace(hdc)",
	func(a Value) Value {
		const bufsize = 512
		var buf [bufsize]byte
		n, _, _ := getTextFace.Call(
			intArg(a),
			bufsize,
			uintptr(unsafe.Pointer(&buf[0])))
		return SuStr(string(buf[:n]))
	})

// dll long Gdi32:SetBkMode(pointer hdc, long mode)
var setBkMode = gdi32.NewProc("SetBkMode")
var _ = builtin2("SetBkMode(hdc, color)",
	func(a, b Value) Value {
		rtn, _, _ := setBkMode.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll long Gdi32:SetBkColor(pointer hdc, long color)
var setBkColor = gdi32.NewProc("SetBkColor")
var _ = builtin2("SetBkColor(hdc, color)",
	func(a, b Value) Value {
		rtn, _, _ := setBkColor.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll Gdi32:DeleteDC(pointer hdc) bool
var deleteDC = gdi32.NewProc("DeleteDC")
var _ = builtin1("DeleteDC(hdc)",
	func(a Value) Value {
		rtn, _, _ := deleteDC.Call(intArg(a))
		return intRet(rtn)
	})

// dll Gdi32:DeleteObject(pointer hgdiobj) bool
var deleteObject = gdi32.NewProc("DeleteObject")
var _ = builtin1("DeleteObject(hgdiobj)", func(a Value) Value {
	rtn, _, _ := deleteObject.Call(intArg(a))
	return boolRet(rtn)
})

// dll Gdi32:GetClipBox(pointer hdc, RECT* rect) long
var getClipBox = gdi32.NewProc("GetClipBox")
var _ = builtin2("GetClipBox(hdc, rect)",
	func(a, b Value) Value {
		var r RECT
		rtn, _, _ := getClipBox.Call(
			intArg(a),
			uintptr(rectArg(b, &r)))
		return intRet(rtn)
	})

// dll Gdi32:GetPixel(pointer hdc, long nXPos, long nYPos) long
var getPixel = gdi32.NewProc("GetPixel")
var _ = builtin3("GetPixel(hdc, nXPos, nYPos)",
	func(a, b, c Value) Value {
		rtn, _, _ := getPixel.Call(
			intArg(a),
			intArg(b),
			intArg(c))
		return intRet(rtn)
	})

// dll Gdi32:PtVisible(pointer hdc, long nXPos, long nYPos) bool
var ptVisible = gdi32.NewProc("PtVisible")
var _ = builtin3("PtVisible(hdc, nXPos, nYPos)",
	func(a, b, c Value) Value {
		rtn, _, _ := ptVisible.Call(
			intArg(a),
			intArg(b),
			intArg(c))
		if rtn == 0xffffffff { // 0xffffffff: 32 bit -1
			return IntVal(-1)
		}
		return boolRet(rtn)
	})

// dll Gdi32:Rectangle(pointer hdc, long left, long top, long right, long bottom) bool
var rectangle = gdi32.NewProc("Rectangle")
var _ = builtin5("Rectangle(hdc, left, top, right, bottom)",
	func(a, b, c, d, e Value) Value {
		rtn, _, _ := rectangle.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			intArg(d),
			intArg(e))
		return intRet(rtn)
	})

// dll Gdi32:SetStretchBltMode(pointer hdc, long iStretchMode) long
var setStretchBltMode = gdi32.NewProc("SetStretchBltMode")
var _ = builtin2("SetStretchBltMode(hdc, iStretchMode)",
	func(a, b Value) Value {
		rtn, _, _ := setStretchBltMode.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll Gdi32:SetTextColor(pointer hdc, long color) long
var setTextColor = gdi32.NewProc("SetTextColor")
var _ = builtin2("SetTextColor(hdc, color)",
	func(a, b Value) Value {
		rtn, _, _ := setTextColor.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll Gdi32:StretchBlt(pointer hdcDest, long nXOriginDest, long nYOriginDest, long nWidthDest, long nHeightDest, pointer hdcSrc,
//	long nXOriginSrc, long nYOriginSrc, long nWidthSrc, long nHeightSrc, long dwRop) bool
var stretchBlt = gdi32.NewProc("StretchBlt")
var _ = builtin("StretchBlt(hdcDest, nXOriginDest, nYOriginDest, nWidthDest, nHeightDest, hdcSrc, nXOriginSrc, nYOriginSrc, nWidthSrc, nHeightSrc, dwRop)",
	func(_ *Thread, a []Value) Value {
		rtn, _, _ := stretchBlt.Call(
			intArg(a[0]),
			intArg(a[1]),
			intArg(a[2]),
			intArg(a[3]),
			intArg(a[4]),
			intArg(a[5]),
			intArg(a[6]),
			intArg(a[7]),
			intArg(a[8]),
			intArg(a[9]),
			intArg(a[10]))
		return boolRet(rtn)
	})

// dll pointer Gdi32:CloseEnhMetaFile(pointer dc)
var closeEnhMetaFile = gdi32.NewProc("CloseEnhMetaFile")
var _ = builtin1("CloseEnhMetaFile(dc)",
	func(a Value) Value {
		rtn, _, _ := closeEnhMetaFile.Call(
			intArg(a))
		return intRet(rtn)
	})

// dll pointer Gdi32:CreateDC(
// 	[in] string driver,
// 	[in] string device,
// 	[in] string output,
// 	pointer devmode)
var createDC = gdi32.NewProc("CreateDC")
var _ = builtin4("CreateDC(driver, device, output, devmode)",
	func(a, b, c, d Value) Value {
		rtn, _, _ := createDC.Call(
			uintptr(stringArg(a)),
			uintptr(stringArg(b)),
			uintptr(stringArg(c)),
			intArg(d))
		return intRet(rtn)
	})

// dll bool Gdi32:Ellipse(pointer hdc, long left, long top, long right, long bottom)
var ellipse = gdi32.NewProc("Ellipse")
var _ = builtin5("Ellipse(hdc, left, top, right, bottom)",
	func(a, b, c, d, e Value) Value {
		rtn, _, _ := ellipse.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			intArg(d),
			intArg(e))
		return boolRet(rtn)
	})

// dll long Gdi32:EndDoc(pointer hdc)
var endDoc = gdi32.NewProc("EndDoc")
var _ = builtin1("EndDoc(hdc)",
	func(a Value) Value {
		rtn, _, _ := endDoc.Call(
			intArg(a))
		return intRet(rtn)
	})

// dll long Gdi32:EndPage(pointer hdc)
var endPage = gdi32.NewProc("EndPage")
var _ = builtin1("EndPage(hdc)",
	func(a Value) Value {
		rtn, _, _ := endPage.Call(
			intArg(a))
		return intRet(rtn)
	})

// dll long Gdi32:ExcludeClipRect(pointer hdc, long l, long t, long r, long b)
var excludeClipRect = gdi32.NewProc("ExcludeClipRect")
var _ = builtin5("ExcludeClipRect(hdc, l, t, r, b)",
	func(a, b, c, d, e Value) Value {
		rtn, _, _ := excludeClipRect.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			intArg(d),
			intArg(e))
		return intRet(rtn)
	})

// dll long Gdi32:GetClipRgn(pointer hdc, pointer hrgn)
var getClipRgn = gdi32.NewProc("GetClipRgn")
var _ = builtin2("GetClipRgn(hdc, hrgn)",
	func(a, b Value) Value {
		rtn, _, _ := getClipRgn.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll pointer Gdi32:GetCurrentObject(pointer hdc, long uObjectType)
var getCurrentObject = gdi32.NewProc("GetCurrentObject")
var _ = builtin2("GetCurrentObject(hdc, uObjectType)",
	func(a, b Value) Value {
		rtn, _, _ := getCurrentObject.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll pointer Gdi32:GetEnhMetaFile(string filename)
var getEnhMetaFile = gdi32.NewProc("GetEnhMetaFile")
var _ = builtin1("GetEnhMetaFile(filename)",
	func(a Value) Value {
		rtn, _, _ := getEnhMetaFile.Call(
			uintptr(stringArg(a)))
		return intRet(rtn)
	})

// dll bool Gdi32:LineTo(pointer hdc, long x, long y)
var lineTo = gdi32.NewProc("LineTo")
var _ = builtin3("LineTo(hdc, x, y)",
	func(a, b, c Value) Value {
		rtn, _, _ := lineTo.Call(
			intArg(a),
			intArg(b),
			intArg(c))
		return boolRet(rtn)
	})

// dll bool Gdi32:PatBlt(
// 	pointer	hdc,        // Destination device context
// 	long	nXLeft,     // x-coordinate of upper left corner of destination rectangle
// 	long	nYLeft,     // y-coordinate of upper left corner of destination rectangle
// 	long	nWidth,     // width of destination rectangle
// 	long	nHeight,    // height of destination rectangle
// 	long	dwRop       // Raster operation
// 	)
var patBlt = gdi32.NewProc("PatBlt")
var _ = builtin6("PatBlt(hdc, nXLeft, nYLeft, nWidth, nHeight, dwRop)",
	func(a, b, c, d, e, f Value) Value {
		rtn, _, _ := patBlt.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			intArg(d),
			intArg(e),
			intArg(f))
		return boolRet(rtn)
	})

// dll bool Gdi32:Polygon(
// 	pointer hdc,		// handle to device context
// 	[in] string lppt,		// array of points
// 	long cCount		// count of points
// 	)
var polygon = gdi32.NewProc("Polygon")
var _ = builtin3("Polygon(hdc, lppt, cCount)",
	func(a, b, c Value) Value {
		rtn, _, _ := polygon.Call(
			intArg(a),
			uintptr(stringArg(b)),
			intArg(c))
		return boolRet(rtn)
	})

// dll bool Gdi32:RestoreDC(
// 	pointer	hdc,        // handle to DC
// 	long	nSavedDC    // restore state returned by SaveDC
// )
var restoreDC = gdi32.NewProc("RestoreDC")
var _ = builtin2("RestoreDC(hdc, nSavedDC)",
	func(a, b Value) Value {
		rtn, _, _ := restoreDC.Call(
			intArg(a),
			intArg(b))
		return boolRet(rtn)
	})

// dll bool Gdi32:RoundRect(pointer hdc, long left, long top, long right, long bottom,
// 	long ellipse_width, long ellipse_height)
var roundRect = gdi32.NewProc("RoundRect")
var _ = builtin7("RoundRect(hdc, left, top, right, bottom, ellipse_width, ellipse_height)",
	func(a, b, c, d, e, f, g Value) Value {
		rtn, _, _ := roundRect.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			intArg(d),
			intArg(e),
			intArg(f),
			intArg(g))
		return boolRet(rtn)
	})

// dll long Gdi32:SaveDC(
// 	pointer hdc	// handle to DC
// )
//
// // corresponding restore function is RestoreDC...
var saveDC = gdi32.NewProc("SaveDC")
var _ = builtin1("SaveDC(hdc)",
	func(a Value) Value {
		rtn, _, _ := saveDC.Call(
			intArg(a))
		return intRet(rtn)
	})

// dll long Gdi32:SelectClipRgn(pointer hdc, pointer hrgn)
var selectClipRgn = gdi32.NewProc("SelectClipRgn")
var _ = builtin2("SelectClipRgn(hdc, hrgn)",
	func(a, b Value) Value {
		rtn, _, _ := selectClipRgn.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll pointer Gdi32:SetEnhMetaFileBits(
// 	long cbBuffer,
// 	[in] string lpData
// 	)
var setEnhMetaFileBits = gdi32.NewProc("SetEnhMetaFileBits")
var _ = builtin2("SetEnhMetaFileBits(cbBuffer, lpData)",
	func(a, b Value) Value {
		rtn, _, _ := setEnhMetaFileBits.Call(
			intArg(a),
			uintptr(stringArg(b)))
		return intRet(rtn)
	})

// dll long Gdi32:SetMapMode(pointer hdc, long mode)
var setMapMode = gdi32.NewProc("SetMapMode")
var _ = builtin2("SetMapMode(hdc, mode)",
	func(a, b Value) Value {
		rtn, _, _ := setMapMode.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll long Gdi32:SetROP2(pointer hdc, long fnDrawMode)
var setROP2 = gdi32.NewProc("SetROP2")
var _ = builtin2("SetROP2(hdc, fnDrawMode)",
	func(a, b Value) Value {
		rtn, _, _ := setROP2.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll long Gdi32:SetTextAlign(pointer hdc, long mode)
var setTextAlign = gdi32.NewProc("SetTextAlign")
var _ = builtin2("SetTextAlign(hdc, mode)",
	func(a, b Value) Value {
		rtn, _, _ := setTextAlign.Call(
			intArg(a),
			intArg(b))
		return intRet(rtn)
	})

// dll long Gdi32:StartPage(pointer hdc)
var startPage = gdi32.NewProc("StartPage")
var _ = builtin1("StartPage(hdc)",
	func(a Value) Value {
		rtn, _, _ := startPage.Call(
			intArg(a))
		return intRet(rtn)
	})

// dll bool Gdi32:TextOut(pointer hdc, long x, long y, [in] string text, long n)
var textOut = gdi32.NewProc("TextOut")
var _ = builtin5("TextOut(hdc, x, y, text, n)",
	func(a, b, c, d, e Value) Value {
		rtn, _, _ := textOut.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(stringArg(d)),
			intArg(e))
		return boolRet(rtn)
	})

// dll bool Gdi32:Arc(pointer hdc, long nLeftRect, long nTopRect,
//		long nRightRect, long nBottomRect, long nXStartArc, long nYStartArc,
//		long nXEndArc, long nYEndArc)
var arc = gdi32.NewProc("Arc")
var _ = builtin("Arc(hdc, nLeftRect, nTopRect, nRightRect, nBottomRect,"+
	"nXStartArc, nYStartArc, nXEndArc, nYEndArc)",
	func(_ *Thread, a []Value) Value {
		rtn, _, _ := arc.Call(
			intArg(a[0]),
			intArg(a[1]),
			intArg(a[2]),
			intArg(a[3]),
			intArg(a[4]),
			intArg(a[5]),
			intArg(a[6]),
			intArg(a[7]),
			intArg(a[8]))
		return intRet(rtn)
	})

// dll pointer Gdi32:CreateEnhMetaFile(pointer hdcRef, [in] string filename,
//		RECT* rect, [in] string desc)
var createEnhMetaFile = gdi32.NewProc("CreateEnhMetaFileA")
var _ = builtin4("CreateEnhMetaFile(hdcRef, filename, rect, desc)",
	func(a, b, c, d Value) Value {
		var r RECT
		rtn, _, _ := createEnhMetaFile.Call(
			intArg(a),
			uintptr(stringArg(b)),
			uintptr(rectArg(c, &r)),
			uintptr(stringArg(d)))
		return intRet(rtn)
	})

// dll gdiobj Gdi32:CreateFont(long nHeight, long nWidth, long nEscapement,
//		long nOrientation, long fnWeight, long fdwItalic, long fdwUnderline,
//		long fdwStrikeOut, long fdwCharSet, long fdwOutputPrecision,
//		long fdwClipPrecision, long fdwQuality, long fdwPitchAndFamily,
//		[in] string lpszFace)
var createFont = gdi32.NewProc("CreateFont")
var _ = builtin("CreateFont(hdc, x, y, cx, cy, hdcSrc, x1, y1, rop)",
	func(_ *Thread, a []Value) Value {
		rtn, _, _ := createFont.Call(
			intArg(a[0]),
			intArg(a[1]),
			intArg(a[2]),
			intArg(a[3]),
			intArg(a[4]),
			intArg(a[5]),
			intArg(a[6]),
			intArg(a[7]),
			intArg(a[8]),
			intArg(a[9]),
			intArg(a[10]),
			intArg(a[11]),
			intArg(a[12]),
			uintptr(stringArg(a[13])))
		return intRet(rtn)
	})

// dll bool Gdi32:ExtTextOut(pointer hdc, long X, long Y, long fuOptions,
//		RECT* lprc, [in] string lpString, long cbCount, LONG* lpDx)
var extTextOut = gdi32.NewProc("ExtTextOutA")
var _ = builtin("ExtTextOut(hdc, x, y, fuOptions, lprc, lpString, cbCount,"+
	" lpDx/*unused*/)",
	func(_ *Thread, a []Value) Value {
		var r RECT
		verify.That(a[7].Equal(Zero))
		rtn, _, _ := extTextOut.Call(
			intArg(a[0]),
			intArg(a[1]),
			intArg(a[2]),
			intArg(a[3]),
			uintptr(rectArg(a[4], &r)),
			uintptr(stringArg(a[5])),
			intArg(a[6]),
			0)
		return intRet(rtn)
	})

// dll gdiobj Gdi32:ExtCreatePen(long dwPenStyle, long dwWidth, LOGBRUSH* brush,
//		long dwStyleCount, pointer lpStyle)
var extCreatePen = gdi32.NewProc("ExtCreatePen")
var _ = builtin5("ExtCreatePen(dwPenStyle, dwWidth, brush, "+
	"dwStyleCount/*unused*/, lpStyle/*unused*/)",
	func(a, b, c, d, e Value) Value {
		lb := LOGBRUSH{
			lbStyle: getInt32(c, "lbStyle"),
			lbColor: getInt32(c, "lbColor"),
			lbHatch: uintptr(getInt(d, "dwItemData")),
		}
		rtn, _, _ := extCreatePen.Call(
			intArg(a),
			intArg(b),
			uintptr(unsafe.Pointer(&lb)),
			0,
			0)
		return intRet(rtn)
	})

type LOGBRUSH struct {
	lbStyle int32
	lbColor int32
	lbHatch HANDLE
}

// dll long Gdi32:GetGlyphOutline(pointer hdc, long uChar, long uFormat,
//		GLYPHMETRICS*  lpgm, long cbBuffer, pointer lpvBuffer, MAT2* lpmat2)
var getGlyphOutline = gdi32.NewProc("GetGlyphOutlineA")
var _ = builtin7("GetGlyphOutline(hdc, uChar, uFormat, lpgm, "+
	"cbBuffer/*unused*/, lpvBuffer/*unused*/, lpmat2)",
	func(a, b, c, d, e, f, g Value) Value {
		var gm GLYPHMETRICS
		mat := MAT2{
			eM11: FIXED{
				fract: getInt32(g.Get(nil, SuStr("eM11")), "fract"),
				value: getInt32(g.Get(nil, SuStr("eM11")), "value")},
			eM12: FIXED{
				fract: getInt32(g.Get(nil, SuStr("eM12")), "fract"),
				value: getInt32(g.Get(nil, SuStr("eM12")), "value")},
			eM21: FIXED{
				fract: getInt32(g.Get(nil, SuStr("eM21")), "fract"),
				value: getInt32(g.Get(nil, SuStr("eM21")), "value")},
			eM22: FIXED{
				fract: getInt32(g.Get(nil, SuStr("eM22")), "fract"),
				value: getInt32(g.Get(nil, SuStr("eM22")), "value")},
		}
		rtn, _, _ := getGlyphOutline.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(unsafe.Pointer(&gm)),
			0,
			0,
			uintptr(unsafe.Pointer(&mat)))
		d.Put(nil, SuStr("gmBlackBoxX"), IntVal(int(gm.gmBlackBoxX)))
		d.Put(nil, SuStr("gmBlackBoxY"), IntVal(int(gm.gmBlackBoxY)))
		d.Put(nil, SuStr("gmptGlyphOrigin"),
			pointToOb(&gm.gmptGlyphOrigin, d.Get(nil, SuStr("gmptGlyphOrigin"))))
		d.Put(nil, SuStr("gmCellIncX"), IntVal(int(gm.gmCellIncX)))
		d.Put(nil, SuStr("gmCellIncY"), IntVal(int(gm.gmCellIncY)))
		return intRet(rtn)
	})

type GLYPHMETRICS struct {
	gmBlackBoxX     int32
	gmBlackBoxY     int32
	gmptGlyphOrigin POINT
	gmCellIncX      int32
	gmCellIncY      int32
}

type FIXED struct {
	fract int32
	value int32
}

type MAT2 struct {
	eM11 FIXED
	eM12 FIXED
	eM21 FIXED
	eM22 FIXED
}

// dll long Gdi32:StartDoc(pointer hdc, DOCINFO* di)
var startDoc = gdi32.NewProc("StartDocA")
var _ = builtin2("StartDoc(hdc, di)",
	func(a, b Value) Value {
		di := DOCINFO{
			cbSize:       int32(unsafe.Sizeof(DOCINFO{})),
			lpszDocName:  getStr(b, "lpszDocName"),
			lpszOutput:   getStr(b, "lpszOutput"),
			lpszDatatype: getStr(b, "lpszDatatype"),
			fwType:       getInt32(b, "fwType"),
		}
		rtn, _, _ := startDoc.Call(
			intArg(a),
			uintptr(unsafe.Pointer(&di)))
		return intRet(rtn)
	})

type DOCINFO struct {
	cbSize       int32
	lpszDocName  *byte
	lpszOutput   *byte
	lpszDatatype *byte
	fwType       int32
}

// dll bool Gdi32:SetWindowExtEx(pointer hdc, long x, long y, POINT* p)
var setWindowExtEx = gdi32.NewProc("SetWindowExtEx")
var _ = builtin4("SetWindowExtEx(hdc, x, y, p)",
	func(a, b, c, d Value) Value {
		var pt POINT
		rtn, _, _ := setWindowExtEx.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(unsafe.Pointer(&pt)))
		pointToOb(&pt, d)
		return boolRet(rtn)
	})

// dll bool Gdi32:SetViewportOrgEx(pointer hdc, long x, long y, POINT* p)
var setViewportOrgEx = gdi32.NewProc("SetViewportOrgEx")
var _ = builtin4("SetViewportOrgEx(hdc, x, y, p)",
	func(a, b, c, d Value) Value {
		var pt POINT
		rtn, _, _ := setViewportOrgEx.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(unsafe.Pointer(&pt)))
		pointToOb(&pt, d)
		return boolRet(rtn)
	})

// dll bool Gdi32:SetViewportExtEx(pointer hdc, long x, long y, POINT* p)
var setViewportExtEx = gdi32.NewProc("SetViewportExtEx")
var _ = builtin4("SetViewportExtEx(hdc, x, y, p)",
	func(a, b, c, d Value) Value {
		var pt POINT
		rtn, _, _ := setViewportExtEx.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(unsafe.Pointer(&pt)))
		pointToOb(&pt, d)
		return boolRet(rtn)
	})

// dll bool Gdi32:PlayEnhMetaFile(pointer hdc, pointer hemf, RECT* rect)
var playEnhMetaFile = gdi32.NewProc("PlayEnhMetaFile")
var _ = builtin3("PlayEnhMetaFile(hdc, hemf, rect)",
	func(a, b, c Value) Value {
		var r RECT
		rtn, _, _ := playEnhMetaFile.Call(
			intArg(a),
			intArg(b),
			uintptr(rectArg(c, &r)))
		return boolRet(rtn)
	})

// dll bool Gdi32:MoveToEx(pointer hdc, long x, long y, POINT* p)
var moveToEx = gdi32.NewProc("MoveToEx")
var _ = builtin4("MoveToEx(hdc, x, y, p)",
	func(a, b, c, d Value) Value {
		var pt POINT
		rtn, _, _ := moveToEx.Call(
			intArg(a),
			intArg(b),
			intArg(c),
			uintptr(unsafe.Pointer(&pt)))
		pointToOb(&pt, d)
		return boolRet(rtn)
	})

// dll long Gdi32:GetObject(pointer hgdiobj, long bufsize, buffer buf)
var getObject = gdi32.NewProc("GetObject")
var _ = builtin1("GetObjectBitmap(h)",
	func(a Value) Value {
		var bm BITMAP
		bmSize := unsafe.Sizeof(bm)
		rtn, _, _ := getObject.Call(
			intArg(a),
			bmSize,
			uintptr(unsafe.Pointer(&bm)))
		if rtn != bmSize {
			return False
		}
		ob := NewSuObject()
		ob.Put(nil, SuStr("bmType"), IntVal(int(bm.bmType)))
		ob.Put(nil, SuStr("bmWidth"), IntVal(int(bm.bmWidth)))
		ob.Put(nil, SuStr("bmHeight"), IntVal(int(bm.bmHeight)))
		ob.Put(nil, SuStr("bmWidthBytes"), IntVal(int(bm.bmWidthBytes)))
		ob.Put(nil, SuStr("bmPlanes"), IntVal(int(bm.bmPlanes)))
		ob.Put(nil, SuStr("bmBitsPixel"), IntVal(int(bm.bmBitsPixel)))
		// bmBits not used
		return ob
	})

type BITMAP struct {
	bmType       uint32
	bmWidth      uint32
	bmHeight     uint32
	bmWidthBytes uint32
	bmPlanes     int16
	bmBitsPixel  int16
	bmBits       uintptr
}