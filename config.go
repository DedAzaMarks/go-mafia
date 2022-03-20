package mafia

const (
	RoleInnocent = iota
	RoleMafia
	RoleDetective
	RoleGhost
	RoleNotAssigned
)

const (
	DayTimeDay = iota
	DayTimeNight
)

const MinUsersNum = 4

const MafiaNum = 1

const (
	GetUsers      = "users"
	Broadcast     = "broadcast"
	VoteFinishDay = "nigh"
	Decision      = "decide"
	Accuse        = "accuse"
)
