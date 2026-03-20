package main

// builtinURLPatterns and builtinContentPatterns mirror the hard-coded
// URL_PATTERNS and CONTENT_PATTERNS from the original Node.js jsaudit.js.
// They supplement the patterns generated from the RetireJS database so that
// we achieve parity with the Node version when detecting bundled assets.

var builtinURLPatterns = map[string][]string{
	"jquery":              {`jquery[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jquery-ui":           {`jquery[.-]ui[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"bootstrap":           {`bootstrap[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"lodash":              {`lodash[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"moment":              {`moment[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"angularjs":           {`angular[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"handlebars":          {`handlebars[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"underscore":          {`underscore[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"axios":               {`axios[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"marked":              {`marked[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"highlight.js":        {`highlight[.-]?(?:js)?[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"socket.io":           {`socket\.io[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"select2":             {`select2[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"tinymce":             {`tinymce[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"ckeditor":            {`ckeditor[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"codemirror":          {`codemirror[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"video.js":            {`video[.-]?js[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jquery-fileupload":   {`jquery[.-]fileupload[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"crypto-js":           {`crypto[.-]?js[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"prototype":           {`prototype[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"dompurify":           {`dompurify[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"knockout":            {`knockout[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"vue":                 {`vue[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"datatables":          {`datatables[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"swiper":              {`swiper[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"slick":               {`slick[/-](\d+\.\d+[\d.]*)(\.min)?\.js`, `slick-(\d+\.\d+[\d.]*)/slick`},
	"owl-carousel":        {`owl\.carousel[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"isotope":             {`isotope[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"flexslider":          {`flexslider[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"chosen":              {`chosen[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"colorbox":            {`colorbox[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"leaflet":             {`leaflet[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"leaflet-markercluster": {`leaflet\.markercluster[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jquery-validation":   {`jquery[.-]validate[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jquery-form":         {`jquery[.-]form[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jszip":               {`jszip[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"plupload":            {`plupload[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"ace-editor":          {`ace[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"quill":               {`quill[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"chart.js":            {`chart[.-]?js[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"d3":                  {`d3[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"three.js":            {`three[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"pdf.js":              {`pdf[.-]?js[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"requirejs":           {`require[.-]?js[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"jquery-migrate":      {`jquery[.-]migrate[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"gsap":                {`gsap[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"vue-i18n":            {`vue-i18n[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
	"react":               {`react[.-](\d+\.\d+[\d.]*)(\.production\.min|\.development)?\.js`},
	"nextjs":              {`next[/-](\d+\.\d+[\d.]*)(\.min)?\.js`},
}

var builtinContentPatterns = map[string][]string{
	// jQuery — patterns are intentionally strict to avoid false positives in plugins
	"jquery": {
		`jQuery JavaScript Library v(\d+\.\d+[\d.]*)`,
		`\\* jQuery v(\d+\.\d+[\d.]*)(?:-[\w.-]+)? \|`,
		`\\{version:"(\d+\.\d+[\d.]*)",fn:`,
	},
	// jQuery UI
	"jquery-ui": {
		`jQuery UI - v(\d+\.\d+[\d.]*)`,
		`ui\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Bootstrap
	"bootstrap": {
		`Bootstrap v(\d+\.\d+[\d.]*)`,
		`\\* Bootstrap v(\d+\.\d+[\d.]*)`,
	},
	// AngularJS
	"angularjs": {
		`AngularJS v(\d+\.\d+[\d.]*)`,
		`angular[:\s]+["'](\d+\.\d+[\d.]*)["']`,
	},
	// Lodash
	"lodash": {
		`lodash (\d+\.\d+[\d.]*)`,
		`Lodash <https://lodash\.com/> (\d+\.\d+[\d.]*)`,
	},
	// Moment
	"moment": {
		`moment\.js.*?(\d+\.\d+[\d.]*)`,
		`\\* @version (\d+\.\d+[\d.]*)[\s\S]{0,20}moment`,
	},
	// Handlebars
	"handlebars": {
		`Handlebars\.VERSION\s*=\s*"(\d+\.\d+[\d.]*)"`,
		`handlebars v(\d+\.\d+[\d.]*)`,
	},
	// Underscore
	"underscore": {
		`Underscore\.js (\d+\.\d+[\d.]*)`,
		`_.VERSION\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Axios
	"axios": {
		`axios\/(\d+\.\d+[\d.]*)`,
		`axios v(\d+\.\d+[\d.]*)`,
	},
	// Marked
	"marked": {
		`marked v(\d+\.\d+[\d.]*)`,
		`marked@(\d+\.\d+[\d.]*)`,
	},
	// Highlight.js
	"highlight.js": {
		`highlight\.js v(\d+\.\d+[\d.]*)`,
		`highlight\.js\s+(\d+\.\d+[\d.]*)`,
	},
	// Socket.IO
	"socket.io": {
		`socket\.io v(\d+\.\d+[\d.]*)`,
		`socket\.io@(\d+\.\d+[\d.]*)`,
	},
	// Select2
	"select2": {
		`Select2 (\d+\.\d+[\d.]*)`,
	},
	// TinyMCE
	"tinymce": {
		`tinymce.*?(\d+\.\d+[\d.]*)`,
		`TinyMCE (\d+\.\d+[\d.]*)`,
	},
	// CKEditor
	"ckeditor": {
		`CKEditor[\s\S]{0,20}(\d+\.\d+[\d.]*)`,
		`CKEDITOR\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// CodeMirror
	"codemirror": {
		`CodeMirror[\s\S]{0,20}(\d+\.\d+[\d.]*)`,
		`CodeMirror\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Video.js
	"video.js": {
		`video\.js\s+(\d+\.\d+[\d.]*)`,
		`videojs\/(\d+\.\d+[\d.]*)`,
		`"Video\.js"[^"]{0,60}"(\d+\.\d+[\d.]*)"`,
		`videojs\.VERSION\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
	},
	// CryptoJS
	"crypto-js": {
		`CryptoJS v(\d+\.\d+[\d.]*)`,
	},
	// Prototype.js
	"prototype": {
		`Prototype JavaScript framework, version (\d+\.\d+[\d.]*)`,
	},
	// DOMPurify
	"dompurify": {
		`DOMPurify[\s\S]{0,20}(\d+\.\d+[\d.]*)`,
	},
	// Knockout
	"knockout": {
		`Knockout JavaScript library v(\d+\.\d+[\d.]*)`,
		`ko\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Vue
	"vue": {
		`Vue\.version\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
		`vue@(\d+\.\d+[\d.]*)`,
	},
	// DataTables
	"datatables": {
		`DataTables (\d+\.\d+[\d.]*)`,
	},
	// Swiper
	"swiper": {
		`Swiper (\d+\.\d+[\d.]*)`,
		`swiper\/(\d+\.\d+[\d.]*)`,
	},
	// Slick
	"slick": {
		`slick - Version (\d+\.\d+[\d.]*)`,
		`slick\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Owl Carousel
	"owl-carousel": {
		`Owl Carousel[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
	},
	// Isotope
	"isotope": {
		`Isotope[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`isotope\.pkgd[\s\S]{0,50}v?(\d+\.\d+[\d.]*)`,
	},
	// FlexSlider
	"flexslider": {
		`FlexSlider[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`jQuery FlexSlider v(\d+\.\d+[\d.]*)`,
	},
	// Chosen
	"chosen": {
		`Chosen[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`chosen\.js.*?(\d+\.\d+[\d.]*)`,
	},
	// Colorbox
	"colorbox": {
		`jQuery Colorbox[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`colorbox.*?v(\d+\.\d+[\d.]*)`,
	},
	// Leaflet
	"leaflet": {
		`Leaflet[\s\S]{0,10}v?(\d+\.\d+[\d.]*)`,
		`L\.version\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
	},
	// Leaflet MarkerCluster
	"leaflet-markercluster": {
		`Leaflet\.markercluster[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
	},
	// jQuery Validation
	"jquery-validation": {
		`jQuery Validation Plugin[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`\\$.validator\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// jQuery Form
	"jquery-form": {
		`jQuery Form Plugin[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
	},
	// JSZip
	"jszip": {
		`JSZip[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`JSZip\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Plupload
	"plupload": {
		`plupload[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
	},
	// Ace Editor
	"ace-editor": {
		`Ace[\s\S]{0,20}version\s*["'](\d+\.\d+[\d.]*)["']`,
	},
	// Quill
	"quill": {
		`Quill Editor[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`Quill\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// Chart.js
	"chart.js": {
		`Chart\.js[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`Chart\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// D3
	"d3": {
		`d3\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
		`D3[\s\S]{0,10}v(\d+\.\d+[\d.]*)`,
	},
	// Three.js
	"three.js": {
		`THREE\.REVISION\s*=\s*"(\d+)"`,
	},
	// PDF.js
	"pdf.js": {
		`pdfjsLib\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
		`PDF\.js v(\d+\.\d+[\d.]*)`,
	},
	// RequireJS
	"requirejs": {
		`RequireJS[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`require\.version\s*=\s*"(\d+\.\d+[\d.]*)"`,
	},
	// jQuery Migrate
	"jquery-migrate": {
		`jQuery Migrate[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
	},
	// GSAP — bundle-embedded detection
	"gsap": {
		`gsap\.version\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
		`"gsap",\s*"version",\s*"(\d+\.\d+[\d.]*)"`,
		`GSAP[^"]{0,30}["'](\d+\.\d+[\d.]*)["']`,
		`gsap[\s\S]{0,30}version:\s*["'](\d+\.\d+[\d.]*)["']`,
	},
	// Vue I18n / intlify
	"vue-i18n": {
		`intlify[\s\S]{0,30}version:\s*["'](\d+\.\d+[\d.]*)["']`,
		`@intlify\/core-base[\s\S]{0,30}["'](\d+\.\d+[\d.]*)["']`,
		`vue-i18n v(\d+\.\d+[\d.]*)`,
	},
	// React — production bundles embed version as string
	"react": {
		`react\.version\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
		`["']react["'][^"']{0,20}["'](\d+\.\d+[\d.]*)["']`,
		`\bReact\s*\.version\s*=\s*["'](\d+\.\d+[\d.]*)["']`,
	},
	// Next.js
	"nextjs": {
		`next\.js[\s\S]{0,20}v?(\d+\.\d+[\d.]*)`,
		`"next"\s*:\s*"(\d+\.\d+[\d.]*)"`,
		`__NEXT_VERSION[^"']{0,10}["'](\d+\.\d+[\d.]*)["']`,
	},
}

// mergeBuiltinPatterns appends the built-in patterns to those generated from
// the RetireJS DB so that detection matches the Node.js implementation.
func mergeBuiltinPatterns(d *db) {
	for lib, pats := range builtinURLPatterns {
		d.URLPatterns[lib] = append(d.URLPatterns[lib], pats...)
	}
	for lib, pats := range builtinContentPatterns {
		d.ContentRegex[lib] = append(d.ContentRegex[lib], pats...)
	}
}

