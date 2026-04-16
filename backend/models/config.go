package models

// FullConfig 对应 settings.json 的根结构
type FullConfig struct {
	SiteConfig                SiteConfig           `json:"siteConfig"`
	FullscreenWallpaperConfig WallpaperConfig      `json:"fullscreenWallpaperConfig"`
	NavBarConfig              NavBarConfig         `json:"navBarConfig"`
	ProfileConfig             ProfileConfig        `json:"profileConfig"`
	LicenseConfig             LicenseConfig        `json:"licenseConfig"`
	ExpressiveCodeConfig      ExpressiveCodeConfig `json:"expressiveCodeConfig"`
	CommentConfig             CommentConfig        `json:"commentConfig"`
	AnnouncementConfig        AnnouncementConfig   `json:"announcementConfig"`
	MusicPlayerConfig         MusicPlayerConfig    `json:"musicPlayerConfig"`
	FooterConfig              FooterConfig         `json:"footerConfig"`
	SidebarLayoutConfig       SidebarLayoutConfig  `json:"sidebarLayoutConfig"`
	SakuraConfig              SakuraConfig         `json:"sakuraConfig"`
	PioConfig                 PioConfig            `json:"pioConfig"`
}

// ---------- SiteConfig ----------
type SiteConfig struct {
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle"`
	SiteURL          string          `json:"siteURL"`
	SiteStartDate    string          `json:"siteStartDate"`
	TimeZone         int             `json:"timeZone"`
	Lang             string          `json:"lang"`
	ThemeColor       ThemeColor      `json:"themeColor"`
	FeaturePages     map[string]bool `json:"featurePages"`
	NavbarTitle      NavbarTitle     `json:"navbarTitle"`
	Bangumi          Bangumi         `json:"bangumi"`
	Anime            Anime           `json:"anime"`
	PostListLayout   PostListLayout  `json:"postListLayout"`
	TagStyle         TagStyle        `json:"tagStyle"`
	WallpaperMode    WallpaperMode   `json:"wallpaperMode"`
	Banner           Banner          `json:"banner"`
	TOC              TOC             `json:"toc"`
	GenerateOgImages bool            `json:"generateOgImages"`
	Favicon          []interface{}   `json:"favicon"`
	Font             Font            `json:"font"`
	ShowLastModified bool            `json:"showLastModified"`
}

type ThemeColor struct {
	Hue   int  `json:"hue"`
	Fixed bool `json:"fixed"`
}

type NavbarTitle struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type Bangumi struct {
	UserID string `json:"userId"`
}

type Anime struct {
	Mode string `json:"mode"`
}

type PostListLayout struct {
	DefaultMode string `json:"defaultMode"`
	AllowSwitch bool   `json:"allowSwitch"`
}

type TagStyle struct {
	UseNewStyle bool `json:"useNewStyle"`
}

type WallpaperMode struct {
	DefaultMode            string `json:"defaultMode"`
	ShowModeSwitchOnMobile string `json:"showModeSwitchOnMobile"`
}

type Banner struct {
	Src      ImageSrc    `json:"src"`
	Position string      `json:"position"`
	Carousel Carousel    `json:"carousel"`
	Waves    Waves       `json:"waves"`
	ImageAPI ImageAPI    `json:"imageApi"`
	HomeText HomeText    `json:"homeText"`
	Credit   Credit      `json:"credit"`
	Navbar   NavbarTrans `json:"navbar"`
}

type ImageSrc struct {
	Desktop []string `json:"desktop"`
	Mobile  []string `json:"mobile"`
}

type Carousel struct {
	Enable   bool    `json:"enable"`
	Interval float64 `json:"interval"`
}

type Waves struct {
	Enable          bool `json:"enable"`
	PerformanceMode bool `json:"performanceMode"`
	MobileDisable   bool `json:"mobileDisable"`
}

type ImageAPI struct {
	Enable bool   `json:"enable"`
	URL    string `json:"url"`
}

type HomeText struct {
	Enable     bool       `json:"enable"`
	Title      string     `json:"title"`
	Subtitle   []string   `json:"subtitle"`
	Typewriter Typewriter `json:"typewriter"`
}

type Typewriter struct {
	Enable      bool `json:"enable"`
	Speed       int  `json:"speed"`
	DeleteSpeed int  `json:"deleteSpeed"`
	PauseTime   int  `json:"pauseTime"`
}

type Credit struct {
	Enable bool   `json:"enable"`
	Text   string `json:"text"`
	URL    string `json:"url"`
}

type NavbarTrans struct {
	TransparentMode string `json:"transparentMode"`
}

type TOC struct {
	Enable           bool `json:"enable"`
	Depth            int  `json:"depth"`
	UseJapaneseBadge bool `json:"useJapaneseBadge"`
}

type Font struct {
	ASCIIFont FontConfig `json:"asciiFont"`
	CJKFont   FontConfig `json:"cjkFont"`
}

type FontConfig struct {
	FontFamily     string   `json:"fontFamily"`
	FontWeight     string   `json:"fontWeight"`
	LocalFonts     []string `json:"localFonts"`
	EnableCompress bool     `json:"enableCompress"`
}

// ---------- WallpaperConfig (全屏壁纸) ----------
type WallpaperConfig struct {
	Src      ImageSrc `json:"src"`
	Position string   `json:"position"`
	Carousel Carousel `json:"carousel"`
	ZIndex   int      `json:"zIndex"`
	Opacity  float64  `json:"opacity"`
	Blur     int      `json:"blur"`
}

// ---------- NavBarConfig ----------
type NavBarConfig struct {
	Links []interface{} `json:"links"` // 使用 interface{} 兼容预设和自定义结构
}

// ---------- ProfileConfig ----------
type ProfileConfig struct {
	Avatar     string         `json:"avatar"`
	Name       string         `json:"name"`
	Bio        string         `json:"bio"`
	Typewriter TypewriterProf `json:"typewriter"`
	Links      []LinkItem     `json:"links"`
}

type TypewriterProf struct {
	Enable bool `json:"enable"`
	Speed  int  `json:"speed"`
}

type LinkItem struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	URL  string `json:"url"`
}

// ---------- LicenseConfig ----------
type LicenseConfig struct {
	Enable bool   `json:"enable"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

// ---------- ExpressiveCodeConfig ----------
type ExpressiveCodeConfig struct {
	Theme                     string `json:"theme"`
	HideDuringThemeTransition bool   `json:"hideDuringThemeTransition"`
}

// ---------- CommentConfig ----------
type CommentConfig struct {
	Enable bool         `json:"enable"`
	Twikoo TwikooConfig `json:"twikoo"`
}

type TwikooConfig struct {
	EnvID string `json:"envId"`
	Lang  string `json:"lang"`
}

// ---------- AnnouncementConfig ----------
type AnnouncementConfig struct {
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Closable bool       `json:"closable"`
	Link     LinkConfig `json:"link"`
}

type LinkConfig struct {
	Enable   bool   `json:"enable"`
	Text     string `json:"text"`
	URL      string `json:"url"`
	External bool   `json:"external"`
}

// ---------- MusicPlayerConfig ----------
type MusicPlayerConfig struct {
	Enable    bool   `json:"enable"`
	Mode      string `json:"mode"`
	MetingAPI string `json:"meting_api"`
	ID        string `json:"id"`
	Server    string `json:"server"`
	Type      string `json:"type"`
}

// ---------- FooterConfig ----------
type FooterConfig struct {
	Enable     bool   `json:"enable"`
	CustomHTML string `json:"customHtml"`
}

// ---------- SidebarLayoutConfig ----------
type SidebarLayoutConfig struct {
	Position         string             `json:"position"`
	Components       []SidebarComponent `json:"components"`
	DefaultAnimation DefaultAnimation   `json:"defaultAnimation"`
	Responsive       ResponsiveConfig   `json:"responsive"`
}

type SidebarComponent struct {
	Type           string               `json:"type"`
	Enable         bool                 `json:"enable"`
	Order          int                  `json:"order"`
	Position       string               `json:"position"`
	Sidebar        string               `json:"sidebar"`
	Class          string               `json:"class"`
	AnimationDelay int                  `json:"animationDelay"`
	Responsive     *ComponentResponsive `json:"responsive,omitempty"`
}

type ComponentResponsive struct {
	CollapseThreshold int `json:"collapseThreshold"`
}

type DefaultAnimation struct {
	Enable    bool `json:"enable"`
	BaseDelay int  `json:"baseDelay"`
	Increment int  `json:"increment"`
}

type ResponsiveConfig struct {
	Breakpoints Breakpoints     `json:"breakpoints"`
	Layout      LayoutBreakdown `json:"layout"`
}

type Breakpoints struct {
	Mobile  int `json:"mobile"`
	Tablet  int `json:"tablet"`
	Desktop int `json:"desktop"`
}

type LayoutBreakdown struct {
	Mobile  string `json:"mobile"`
	Tablet  string `json:"tablet"`
	Desktop string `json:"desktop"`
}

// ---------- SakuraConfig ----------
type SakuraConfig struct {
	Enable     bool        `json:"enable"`
	SakuraNum  int         `json:"sakuraNum"`
	LimitTimes int         `json:"limitTimes"`
	Size       RangeFloat  `json:"size"`
	Opacity    RangeFloat  `json:"opacity"`
	Speed      SakuraSpeed `json:"speed"`
	ZIndex     int         `json:"zIndex"`
}

type RangeFloat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type SakuraSpeed struct {
	Horizontal RangeFloat `json:"horizontal"`
	Vertical   RangeFloat `json:"vertical"`
	Rotation   float64    `json:"rotation"`
	FadeSpeed  float64    `json:"fadeSpeed"`
}

// ---------- PioConfig ----------
type PioConfig struct {
	Enable         bool     `json:"enable"`
	Models         []string `json:"models"`
	Position       string   `json:"position"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	Mode           string   `json:"mode"`
	HiddenOnMobile bool     `json:"hiddenOnMobile"`
	Dialog         Dialog   `json:"dialog"`
}

type Dialog struct {
	Welcome string   `json:"welcome"`
	Touch   []string `json:"touch"`
	Home    string   `json:"home"`
	Skin    []string `json:"skin"`
	Close   string   `json:"close"`
	Link    string   `json:"link"`
}
