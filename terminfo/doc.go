/*Package terminfo contains a simple and incomplete implementation of the
terminfo database. Information was taken from the ncurses manpages term(5) and
terminfo(5). Currently, only the string capabilities for special keys and for
functions without parameters are actually used. Colors are still done with ANSI
escape sequences. Other special features that are not (yet?) supported are
reading from ~/.terminfo, the TERMINFO_DIRS variable, Berkeley database format
and extended capabilities.

It is currently in the process of evolving out of termbox, and will become more
complete over time.

*/
package terminfo
