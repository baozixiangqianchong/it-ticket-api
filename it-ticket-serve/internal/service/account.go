package service

func canHandleTickets(role string) bool {
	return role == "agent" || role == "admin"
}

func shouldReleaseTickets(fromRole, toRole string, disabling bool) bool {
	if disabling {
		return canHandleTickets(fromRole)
	}
	return canHandleTickets(fromRole) && !canHandleTickets(toRole)
}
