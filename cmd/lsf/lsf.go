package lsf

import (
	"fmt"
	"io"
	"os"

	"github.com/ncw/rclone/cmd"
	"github.com/ncw/rclone/cmd/ls/lshelp"
	"github.com/ncw/rclone/fs"
	"github.com/ncw/rclone/fs/hash"
	"github.com/ncw/rclone/fs/operations"
	"github.com/ncw/rclone/fs/walk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	format    string
	separator string
	dirSlash  bool
	recurse   bool
	hashType  = hash.MD5
	filesOnly bool
	dirsOnly  bool
)

func init() {
	cmd.Root.AddCommand(commandDefintion)
	flags := commandDefintion.Flags()
	flags.StringVarP(&format, "format", "F", "p", "Output format - see  help for details")
	flags.StringVarP(&separator, "separator", "s", ";", "Separator for the items in the format.")
	flags.BoolVarP(&dirSlash, "dir-slash", "d", true, "Append a slash to directory names.")
	flags.VarP(&hashType, "hash", "", "Use this hash when `h` is used in the format MD5|SHA-1|DropboxHash")
	flags.BoolVarP(&filesOnly, "files-only", "", false, "Only list files.")
	flags.BoolVarP(&dirsOnly, "dirs-only", "", false, "Only list directories.")
	commandDefintion.Flags().BoolVarP(&recurse, "recursive", "R", false, "Recurse into the listing.")
}

var commandDefintion = &cobra.Command{
	Use:   "lsf remote:path",
	Short: `List directories and objects in remote:path formatted for parsing`,
	Long: `
List the contents of the source path (directories and objects) to
standard output in a form which is easy to parse by scripts.  By
default this will just be the names of the objects and directories,
one per line.  The directories will have a / suffix.

Eg

    $ rclone lsf swift:bucket
    bevajer5jef
    canole
    diwogej7
    ferejej3gux/
    fubuwic

Use the --format option to control what gets listed.  By default this
is just the path, but you can use these parameters to control the
output:

    p - path
    s - size
    t - modification time
    h - hash

So if you wanted the path, size and modification time, you would use
--format "pst", or maybe --format "tsp" to put the path last.

Eg

    $ rclone lsf  --format "tsp" swift:bucket
    2016-06-25 18:55:41;60295;bevajer5jef
    2016-06-25 18:55:43;90613;canole
    2016-06-25 18:55:43;94467;diwogej7
    2018-04-26 08:50:45;0;ferejej3gux/
    2016-06-25 18:55:40;37600;fubuwic

If you specify "h" in the format you will get the MD5 hash by default,
use the "--hash" flag to change which hash you want.  Note that this
can be returned as an empty string if it isn't available on the object
(and for directories), "ERROR" if there was an error reading it from
the object and "UNSUPPORTED" if that object does not support that hash
type.

For example to emulate the md5sum command you can use

    rclone lsf -R --hash MD5 --format hp --separator "  " --files-only .

Eg

    $ rclone lsf -R --hash MD5 --format hp --separator "  " --files-only swift:bucket 
    7908e352297f0f530b84a756f188baa3  bevajer5jef
    cd65ac234e6fea5925974a51cdd865cc  canole
    03b5341b4f234b9d984d03ad076bae91  diwogej7
    8fd37c3810dd660778137ac3a66cc06d  fubuwic
    99713e14a4c4ff553acaf1930fad985b  gixacuh7ku

(Though "rclone md5sum ." is an easier way of typing this.)

By default the separator is ";" this can be changed with the
--separator flag.  Note that separators aren't escaped in the path so
putting it last is a good strategy.

Eg

    $ rclone lsf  --separator "," --format "tshp" swift:bucket
    2016-06-25 18:55:41,60295,7908e352297f0f530b84a756f188baa3,bevajer5jef
    2016-06-25 18:55:43,90613,cd65ac234e6fea5925974a51cdd865cc,canole
    2016-06-25 18:55:43,94467,03b5341b4f234b9d984d03ad076bae91,diwogej7
    2018-04-26 08:52:53,0,,ferejej3gux/
    2016-06-25 18:55:40,37600,8fd37c3810dd660778137ac3a66cc06d,fubuwic

` + lshelp.Help,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fsrc := cmd.NewFsSrc(args)
		cmd.Run(false, false, command, func() error {
			return Lsf(fsrc, os.Stdout)
		})
	},
}

// Lsf lists all the objects in the path with modification time, size
// and path in specific format.
func Lsf(fsrc fs.Fs, out io.Writer) error {
	var list operations.ListFormat
	list.SetSeparator(separator)
	list.SetDirSlash(dirSlash)

	for _, char := range format {
		switch char {
		case 'p':
			list.AddPath()
		case 't':
			list.AddModTime()
		case 's':
			list.AddSize()
		case 'h':
			list.AddHash(hashType)
		default:
			return errors.Errorf("Unknown format character %q", char)
		}
	}

	return walk.Walk(fsrc, "", false, operations.ConfigMaxDepth(recurse), func(path string, entries fs.DirEntries, err error) error {
		if err != nil {
			fs.CountError(err)
			fs.Errorf(path, "error listing: %v", err)
			return nil
		}
		for _, entry := range entries {
			_, isDir := entry.(fs.Directory)
			if isDir {
				if filesOnly {
					continue
				}
			} else {
				if dirsOnly {
					continue
				}
			}
			fmt.Fprintln(out, operations.ListFormatted(&entry, &list))
		}
		return nil
	})
}
