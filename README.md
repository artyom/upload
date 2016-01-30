Command upload uploads files to new directory on a remote server via ssh. For
each command call new randomly-named remote subdirectory is created.

	Usage: upload [flags] file...
	  -addr string
		ssh host:port (default "localhost:22")
	  -dir string
		remote directory to upload files to (default "/tmp")
	  -long
		generate long subdirectory name
	  -url string
		remote url base to open after upload
	  -user string
		ssh connection username (default "$USER")

Current limitations:

 * only files are supported, not directories;
 * ssh-agent is used for authentication.

## Advanced usage

Can be used to create basic one-person web-share service:

Consider you have a cheap VPS serving `/var/www` as http://example.com. Create
directory `/var/www/pub`, then run upload on some files like this:

	upload -addr my-vps:22 -dir /var/www/pub -url http://example.com/pub cats1.gif cats2.gif

This should create randomly-named directory inside `/var/www/pub`, upload both
files into it, then print directory name to stdout (ex. `/var/www/pub/a1`) and
open http://example.com/pub/a1 in your browser so you can check it and share
(your web server is expected to show directory listings for this to work).

For OS X users such things can be integrated into UI with a little help of
Automator. Run Automator and create a new Service that looks like this:

![Automator action screenshot](https://cbxp.in/c0/automator%20screenshot.png)

Adjust script providing your server details, save this as "Share files on web"
and you'll get a nice Finder context menu:

![Finder context menu displaying custom action](https://cbxp.in/c0/finder%20context%20menu.png)

## LICENSE

[MIT](http://opensource.org/licenses/MIT)
