package main

var (
	ismakecompetition        = make(map[int64]int)
	ismakecomptext           = make(map[int64]int)
	ismakecompbutton         = make(map[int64]int)
	ismakecompmembers        = make(map[int64]int)
	ismakecompdate           = make(map[int64]int)
	ismakecompend            = make(map[int64]int)
	ismakecompcreated        = make(map[int64]int)
	ismakecompetitionchan    = make(map[int64]int)
	ismakecompetitionwintext = make(map[int64]int)
	iscompetitionedit        = make(map[int64]int)
	addchannel               = make(map[int64]int)
)

var Config = struct {
	APPName string `default:"competitionbot"`

	Db struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Host     string `default:"localhost"`
		Port     string `default:"3306"`
		Drop     bool   `default:"false"`
	}

	Bot struct {
		Token string
	}
}{}

var Competition = struct {
	Channel string
	Name    string
	Photo   string
	Text    string
	Button  string
	Date    string
	Members string
	Wintext string
}{}

type Competitions struct {
	Id      int
	Owner   int64
	Channel int64
	Name    string
	Photo   string
	Text    string
	Button  string
	Date    string
	Members int
	Wintext string
}

type Channels struct {
	Id           int
	Owner        int64
	Channelid    int64
	Channelname  string
	Channeltitle string
}

type Participans struct {
	Id       int
	Username string
	Chatid   int64
	Cmptid   int
}
