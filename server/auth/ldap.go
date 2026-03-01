package auth

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type LDAPProvider struct {
	server   string
	port     int
	bindDN   string
	bindPass string
	baseDN   string
	filter   string
}

func NewLDAPProvider(server string, port int, bindDN, bindPass, baseDN, filter string) *LDAPProvider {
	return &LDAPProvider{
		server:   server,
		port:     port,
		bindDN:   bindDN,
		bindPass: bindPass,
		baseDN:   baseDN,
		filter:   filter,
	}
}

func (p *LDAPProvider) Authenticate(username, password string) (string, error) {
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", p.server, p.port))
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	err = conn.Bind(p.bindDN, p.bindPass)
	if err != nil {
		return "", fmt.Errorf("bind failed: %w", err)
	}

	searchReq := ldap.NewSearchRequest(
		p.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf(p.filter, username),
		[]string{"dn", "uid", "cn", "mail"},
		nil,
	)

	searchResult, err := conn.Search(searchReq)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return "", fmt.Errorf("user not found")
	}

	userDN := searchResult.Entries[0].DN

	err = conn.Bind(userDN, password)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	userID := searchResult.Entries[0].GetAttributeValue("uid")
	if userID == "" {
		userID = username
	}

	return userID, nil
}

func (p *LDAPProvider) GetUserAttributes(username string) (map[string]string, error) {
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", p.server, p.port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = conn.Bind(p.bindDN, p.bindPass)
	if err != nil {
		return nil, err
	}

	searchReq := ldap.NewSearchRequest(
		p.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf(p.filter, username),
		[]string{"uid", "cn", "mail", "displayName"},
		nil,
	)

	searchResult, err := conn.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	attrs := make(map[string]string)
	for _, attr := range searchResult.Entries[0].Attributes {
		if len(attr.Values) > 0 {
			attrs[attr.Name] = attr.Values[0]
		}
	}

	return attrs, nil
}
