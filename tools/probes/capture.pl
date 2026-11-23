#!/usr/bin/perl
# capture.pl: READ-ONLY capture of one level of a running d2d through its control socket (card toolstrade14).
# Same job and allow-list as capture.py, for a container without python. Writes NOTHING to disk: every reply goes to
# STDOUT as one JSON line per verb. Run on the host (the script is piped on stdin, the output stays on the host):
#   docker exec -i CONTAINER perl - /run/d2d/d2d.sock.ctl 28 2700 6 < tools/probes/capture.pl > barracks-capture.jsonl
# Args: SOCK [AREA=28 [WAIT_S=2700 [GRACE_S=6]]]. Waits (polling `debug level`) until the level is AREA, sleeps GRACE_S,
# then asks each read verb once. Exit 2 when the area did not show within WAIT_S.
# Env CAPTURE_EXTRA (test only): one more verb to ask, refused unless it is on the allow-list.
use strict; use warnings;
use IO::Socket::UNIX; use IO::Select;

my @READ = ("debug level", "debug levelmap", "debug levelmonsters", "debug levelobjects", "debug presets", "debug objects",
            "debug warps", "debug warptriggers", "debug nearby", "debug layout", "debug grid 5", "state", "info");
sub allowed { my $l = shift; for my $v (@READ, "debug blocked") { return 1 if $l eq $v || index($l, "$v ") == 0 } 0 }
sub jstr { my $s = shift; $s =~ s/([\\"])/\\$1/g; $s =~ s/\n/\\n/g; $s =~ s/\r/\\r/g; $s =~ s/\t/\\t/g;
           $s =~ s/([\x00-\x1f])/sprintf("\\u%04x", ord $1)/ge; "\"$s\"" }

sub ask {
    my ($path, $line) = @_;
    return undef unless allowed($line);
    my $s = IO::Socket::UNIX->new(Peer => $path, Type => SOCK_STREAM) or return "capture error: connect: $!";
    print $s "$line\n";
    my ($buf, $sel) = ("", IO::Select->new($s));
    my $end = time + 40;
    while (time < $end && $sel->can_read(1)) {
        my $n = sysread($s, my $chunk, 1 << 20);
        last unless $n;
        $buf .= $chunk;
        (my $t = $buf) =~ s/\s+$//;
        my ($last) = $t =~ /([^\n]*)\z/;
        last if $last eq "ok" || $last =~ /^ok / || $last =~ /^err/;
    }
    close $s;
    return $buf;
}
sub level_of { my $r = shift; return $1 if defined $r && $r =~ /^ok\s+(\d+)\b/; undef }

my ($sock, $area, $wait, $grace) = ($ARGV[0], $ARGV[1] // 28, $ARGV[2] // 2700, $ARGV[3] // 6);
die "usage: perl - SOCK [AREA [WAIT_S [GRACE_S]]]\n" unless $sock;
$| = 1;
my $t0 = time;
my $seen = 0;
while (time - $t0 < $wait) {
    my $l = level_of(ask($sock, "debug level"));
    if (defined $l && $l == $area) { $seen = 1; last }
    sleep 2;
}
unless ($seen) { print "{\"error\":\"area $area not seen within ${wait}s\"}\n"; exit 2 }
sleep $grace;
my @verbs = (@READ);
push @verbs, $ENV{CAPTURE_EXTRA} if defined $ENV{CAPTURE_EXTRA};
for my $v (@verbs) {
    my $r = ask($sock, $v);
    print "{\"verb\":", jstr($v), ",\"reply\":", (defined $r ? jstr($r) : "null"), ",\"refused\":", (defined $r ? "false" : "true"), "}\n";
}
