# CLOSED source candidate. Never evaluate before OTHER-native and durable root.
# Dedicated external timer, independent of Python/auth/parser/IO workers. It
# enforces real signals; Perl/OS scheduling latency is NOT a hard-real-time proof.
use strict;
use warnings;
use Time::HiRes qw(clock_gettime alarm);
our $ORIGIN;
# Darwin CLOCK_UPTIME_RAW=8 is mach_absolute_time, matching CPython 3.14
# monotonic. Never substitute CLOCK_MONOTONIC (different Darwin epoch).
BEGIN { $ORIGIN = Time::HiRes::clock_gettime(8); }
use POSIX qw(setsid WNOHANG _exit);
use Fcntl qw(F_GETFL F_SETFL O_NONBLOCK);
my ($case,$variant,$release)=@ARGV;
exit 1 unless @ARGV==3 && defined($case) && $case =~ /\A(?:M(?:0[1-9]|1[0-6])|J(?:0[1-9]|1[0-2])|P(?:0[1-9]|1[0-4])|D0[1-6])\z/ && $variant =~ /\A[a-z0-9-]{1,80}\z/ && $release =~ m{\A/};
my $host = $variant =~ /\Ahost-/;
my ($whole,$term,$kill)= $host ? (150,145,149) : (120,115,119);
exit 1 unless -p STDOUT && -p STDERR; # root must capture pipes, no TTY/file/network sink.
my $child;
my $reaped=0;
my $forced=0;
my $limiter=$$;
sub signal_child {
 my ($sig)=@_;
 return unless defined($child) && !$reaped;
 # Only our unreaped direct child can authorize this numeric signal. Before
 # setsid completes, direct PID fallback still reaches the owned fork child.
 kill $sig,-$child;
 kill $sig,$child;
}
$SIG{ALRM}=sub {signal_child('KILL');kill 'KILL',$$;POSIX::_exit(124);};
$SIG{TERM}=sub {$forced=1;signal_child('TERM');};
$SIG{INT}=sub {$forced=1;signal_child('TERM');};
my $remaining=$ORIGIN+$whole-clock_gettime(8);
exit 124 if $remaining<=0;
alarm($remaining); # kernel-backed independent wall timer remains armed through receipt/exit.
$child=fork();
exit 1 unless defined($child);
if($child==0){
 $SIG{ALRM}='DEFAULT';alarm(0);
 POSIX::setsid() >= 0 or POSIX::_exit(126);
 $ENV{D4_NATIVE_LIMITER}="$limiter";
 exec('/opt/homebrew/opt/python@3.14/bin/python3.14','-I','-B','/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-D4-synthetic-harness-preparation-5/outer_host.py','--limited-owner',"$ORIGIN",$case,$variant,$release,"$limiter") or POSIX::_exit(126);
}
# PID direct ownership exists immediately after fork, before metadata or IO.
my $last_term=-1;
my $last_kill=-1;
my $status;
while(1){
 my $elapsed=clock_gettime(8)-$ORIGIN;
 if($elapsed >= $term && $elapsed-$last_term>=.2){$forced=1;signal_child('TERM');$last_term=$elapsed;}
 if($elapsed >= $kill && $elapsed-$last_kill>=.2){$forced=1;signal_child('KILL');$last_kill=$elapsed;}
 if($elapsed >= $whole){signal_child('KILL');POSIX::_exit(124);}
 my $pid=waitpid($child,WNOHANG);
 if($pid==$child){$status=$?;$reaped=1;last;}
 if($pid==-1){$forced=1;signal_child('KILL');POSIX::_exit(125);}
 select undef,undef,undef,.01; # no file/network/hash/parser or blocking wait in timer process.
}
my $elapsed=clock_gettime(8)-$ORIGIN;
my $exit=($status & 127) ? 128+($status & 127) : $status >> 8;
my $ok=(!$forced && $exit==0 && $elapsed<$term)?1:0;
my $flags=fcntl(STDOUT,F_GETFL,0);
POSIX::_exit(125) unless defined($flags) && fcntl(STDOUT,F_SETFL,$flags|O_NONBLOCK);
my $line='{"event":"native-limiter-terminal","limiterPid":'.$limiter.',"ownerPid":'.$child.',"rawOwnerExit":'.$exit.',"elapsedSeconds":'.$elapsed.',"forced":'.($forced?'true':'false').',"localStatus":'.($ok?'"SETTLED_PENDING_ROOT_AUDIT"':'"FAILED"').',"syntheticAcceptance":false,"requiresFreshRootGroupAudit":true}' . "\n";
my $count=syswrite(STDOUT,$line);
POSIX::_exit(125) unless defined($count) && $count==length($line);
POSIX::_exit(1) if clock_gettime(8)-$ORIGIN >= $term;
POSIX::_exit($ok?0:1); # no alarm cancellation or grace after receipt.
