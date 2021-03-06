package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/avahowell/masterkey/repl"
	"github.com/avahowell/masterkey/secureclip"
	"github.com/avahowell/masterkey/vault"
)

var (
	listCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "list",
			Action: list(v),
			Usage:  "list: list the credentials stored inside this vault",
		}
	}

	saveCmd = func(v *vault.Vault, vaultPath string) repl.Command {
		return repl.Command{
			Name:   "save",
			Action: save(v, vaultPath),
			Usage:  "save: save the changes in this vault to disk",
		}
	}

	getCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "get",
			Action: get(v),
			Usage:  "get [location]: get the credential at [location]. [location] can be a partial string: masterkey will search the vault and return the first result.",
		}
	}

	addCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "add",
			Action: add(v),
			Usage:  "add [location] [username] [password]: add a credential to the vault",
		}
	}

	genCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "gen",
			Action: gen(v),
			Usage:  "gen [location] [username]: generate a password and add it to the vault",
		}
	}

	editCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "edit",
			Action: edit(v),
			Usage:  "edit [location] [username] [password]: change the credentials at location to username, password",
		}
	}

	clipCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "clip",
			Action: clip(v),
			Usage:  "clip [location] [meta name]: copy the password at location to the clipboard. meta name optional. Location and meta names can be partial strings, masterkey will search the vault and return the first result.",
		}
	}

	searchCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "search",
			Action: search(v),
			Usage:  "search [searchtext]: search the vault for locations containing searchtext",
		}
	}

	deleteCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "delete",
			Action: deletelocation(v),
			Usage:  "delete [location]: remove [location] from the vault.",
		}
	}
	addmetaCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "addmeta",
			Action: addmeta(v),
			Usage:  "addmeta [location] [meta name] [meta value]: add a metadata tag to the credential at [location]",
		}
	}

	editmetaCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "editmeta",
			Action: editmeta(v),
			Usage:  "editmeta [location] [meta name] [new meta value]: edit an existing metadata tag at [location].",
		}
	}

	deletemetaCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "deletemeta",
			Action: deletemeta(v),
			Usage:  "deletemeta [location] [meta name]: delete an existing metadata tag at [location].",
		}
	}

	importCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "importcsv",
			Action: importcsv(v),
			Usage:  "importcsv [path to csv] [location key] [username key] [password key]: import a csv file.\nThe location key, username key, and password key are the CSV key names used to locate each value. Extra keys will be added to the vault as meta tags.",
		}
	}

	changePasswordCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "changepassword",
			Action: changepassword(v),
			Usage:  "changepassword: change the master password for the vault",
		}
	}

	mergeCmd = func(v *vault.Vault) repl.Command {
		return repl.Command{
			Name:   "merge",
			Action: merge(v),
			Usage:  "merge [location]: merge the vault at location with the currently open vault.",
		}
	}
)

func merge(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("merge requires one argument, the path of the vault to merge")
		}
		vaultPath := args[0]
		pass, err := askPassword("Enter the password for the vault to be merged: ")
		if err != nil {
			return "", err
		}

		vmerge, err := vault.Open(vaultPath, pass)
		if err != nil {
			return "", err
		}
		defer vmerge.Close()

		err = v.Merge(vmerge)
		if err != nil {
			return "", err
		}
		return "vault merged successfully.", nil
	}
}

func changepassword(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		pass1, err := askPassword("Enter a new password for this vault: ")
		if err != nil {
			return "", err
		}
		pass2, err := askPassword("Again, please: ")
		if err != nil {
			return "", err
		}

		if pass1 != pass2 {
			return "", fmt.Errorf("passwords did not match")
		}

		err = v.ChangePassphrase(pass1)
		if err != nil {
			return "", err
		}

		return "master password changed successfully\n", nil
	}
}

func importcsv(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 4 {
			return "", fmt.Errorf("importcsv requires 4 arguments. See help for usage.")
		}

		filepath := args[0]
		locationkey := args[1]
		usernamekey := args[2]
		passwordkey := args[3]

		f, err := os.Open(filepath)
		if err != nil {
			return "", err
		}
		defer f.Close()

		n, err := v.LoadCSV(f, locationkey, usernamekey, passwordkey)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v migrated successfully. %v locations imported.\n", filepath, n), nil
	}
}

func deletemeta(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 2 {
			return "", fmt.Errorf("deletemeta requires 2 arguments. See help for usage.")
		}

		location := args[0]
		metaname := args[1]

		err := v.DeleteMeta(location, metaname)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v deleted from %v successfully.\n", metaname, location), nil
	}
}

func deletelocation(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("delete requires 1 argument. See help for usage.")
		}

		location := args[0]

		err := v.Delete(location)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v deleted successfully.\n", location), nil
	}
}

func editmeta(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 3 {
			return "", fmt.Errorf("editmeta requires 3 arguments. See help for usage.")
		}
		location := args[0]
		metaname := args[1]
		metaval := args[2]

		if err := v.EditMeta(location, metaname, metaval); err != nil {
			return "", err
		}

		return fmt.Sprintf("%v updated successfully.\n", metaname), nil
	}
}

func addmeta(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 3 {
			return "", fmt.Errorf("addmeta requires 3 arguments. See help for usage.")
		}
		location := args[0]
		metaname := args[1]
		metaval := args[2]

		if err := v.AddMeta(location, metaname, metaval); err != nil {
			return "", err
		}

		return fmt.Sprintf("%v added to %v successfully.\n", metaname, location), nil
	}
}

func search(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("search requires 1 argument. See help for usage.")
		}
		searchtext := args[0]

		locations, err := v.Locations()
		if err != nil {
			return "", err
		}

		printstring := ""
		for _, location := range locations {
			if strings.Contains(location, searchtext) {
				printstring += location + "\n"
			}
		}
		return printstring, nil
	}
}

func clip(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) < 1 {
			return "", fmt.Errorf("clip requires at least 1 argument. See help for usage.")
		}

		location, cred, err := v.Find(args[0])
		if err != nil {
			return "", err
		}

		toClip := cred.Password
		clipLabel := cred.Username
		if len(args) > 1 {
			meta := args[1]
			metaname, metaval, err := v.FindMeta(location, meta)
			if err != nil {
				return "", err
			}
			toClip = metaval
			clipLabel = metaname
		}

		err = secureclip.Clip(toClip)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v@%v copied to clipboard, will clear in 30 seconds\n", clipLabel, location), nil
	}
}

func edit(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 3 {
			return "", fmt.Errorf("edit requires 3 arguments. See help for usage.")
		}
		location := args[0]
		credential := vault.Credential{
			Username: args[1],
			Password: args[2],
		}

		err := v.Edit(location, credential)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v updated successfully\n", location), nil
	}
}

func list(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		locations, err := v.Locations()
		if err != nil {
			return "", err
		}
		printstring := "Locations stored in this vault: \n"
		for _, loc := range locations {
			printstring += loc + "\n"
		}
		return printstring, nil
	}
}

func save(v *vault.Vault, savePath string) repl.ActionFunc {
	return func(args []string) (string, error) {
		if err := v.Save(savePath); err != nil {
			return "", err
		}
		return fmt.Sprintf("%v saved successfully.\n", savePath), nil
	}
}

func get(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("get requires at least one argument. See help for usage.")
		}
		_, cred, err := v.Find(args[0])
		if err != nil {
			return "", err
		}

		printstring := fmt.Sprintf("Username: %v\nPassword: %v\n", cred.Username, cred.Password)

		if len(cred.Meta) > 0 {
			for metaname, metaval := range cred.Meta {
				printstring += fmt.Sprintf("%v: %v\n", metaname, metaval)
			}
		}

		return printstring, nil
	}
}

func add(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 3 {
			return "", fmt.Errorf("add requires at least three arguments. See help for usage.")
		}
		location := args[0]
		username := args[1]
		password := args[2]
		cred := vault.Credential{
			Username: username,
			Password: password,
		}
		err := v.Add(location, cred)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v added successfully\n", location), nil
	}
}

func gen(v *vault.Vault) repl.ActionFunc {
	return func(args []string) (string, error) {
		if len(args) != 2 {
			return "", fmt.Errorf("gen requires two arguments. See help for usage.")
		}

		location := args[0]
		username := args[1]

		if err := v.Generate(location, username); err != nil {
			return "", err
		}

		return fmt.Sprintf("%v generated successfully\n", location), nil
	}
}
