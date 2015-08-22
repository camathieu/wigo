package Wigo::Probe;

use strict;
use warnings;

use Getopt::Long;
use JSON;
use File::Basename;

require Exporter;
our @ISA = qw/Exporter/;
our @EXPORT_OK = qw/init config args result version status value message metrics add_metric details raise persist output debug/;
our %EXPORT_TAGS = ( all => \@EXPORT_OK );

###
# VARS
###

my $CONFIG_PATH     = $ENV{"WIGO_PROBE_CONFIG_ROOT"} || "/etc/wigo/conf.d";
my $PERSIST_PATH    = "/tmp";

my $version    = "0.10";

my  $name       = basename($0);
$name =~ s/.pl$//;

my  $config     = {};
my  $args       = [];
my  $persist    = undef;

my  $result     =  {
    version     => "0.10",

    status      => 100,
    level       => "OK",
    message     => "",

    details     => {},
    metrics     => [],
};

###
# COMMAND LINE OPTIONS
###

my $opts = {};
GetOptions (
    $opts,
    'debug',
    '<>' => sub { push @$args, $_[0] }
) or die("Error in command line arguments\n");

my $json = JSON->new;
if ( exists $opts->{'debug'} )
{
    $json = JSON->new->pretty;
}

###
# DEBUG
###

sub debug {
    if ( exists $opts->{'debug'} )
    {
        print shift;
    }
}

###
# OUTPUT JSON
###

sub output {
    my $code = shift;

    $result->{'level'} = getLevel($result->{'status'});

    for my $metric ( @{$result->{'metrics'}} )
    {
        defined $metric->{'Value'} and $metric->{'Value'} += 0;
    }

    print $json->encode( $result ) . "\n";

    if ( defined $code )
    {
        exit $code;
    }
}

###
# GETTER / SETTERS
###

sub init {
    my %params = @_;

    load_config($params{'config'});
    restore();
}

sub config
{
    return $config;
}

sub args
{
    return $args;
}

sub result
{
    return $result;
}

sub version
{
    if ( $@ )
    {
        $result->{"version"} = shift;
    }
    else
    {
        return $result->{"version"};
    }
}

sub status
{
    if ( @_ )
    {
        $result->{"status"} = shift;
    }
    else
    {
        return $result->{"status"};
    }
}

sub message
{
    if ( @_ )
    {
        $result->{"message"} = shift;
    }
    else
    {
        return $result->{"message"};
    }
}

sub metrics
{
    if ( @_ )
    {
        $result->{"metrics"} = shift;
    }
    else
    {
        return $result->{"metrics"};
    }
}

sub add_metric
{
    push @{$result->{"metrics"}}, shift;
}

sub details
{
    if ( @_ )
    {
        $result->{"details"} = shift;
    }
    else
    {
        return $result->{"details"};
    }
}

sub persist
{
    if ( @_ )
    {
        $persist = shift;
        save();
    }
    else
    {
        return $persist;
    }
}

sub raise {
    my $status  = shift;

    result->{'status'} = $status if result->{'status'} < $status;
}

###
# CONFIG
###

sub save_config
{
    my $json = JSON->new->pretty;

    my $path = shift || $CONFIG_PATH . "/" . $name . ".conf";

    if ( open CONFIG, '>', $path )
        {
            eval {
                print CONFIG $json->encode($config)."\n";
            };
            close CONFIG;
    
            if ( $@ )
            {
                status 300;
                message "can't serialize config : $@";
                output 1;
            }
        }
        else
        {
            status 300;
            message "can't open config file $path for writing : $!";
            output 1;
        }
    
}

sub load_config
{
    my $path = $CONFIG_PATH . "/" . $name . ".conf";

    if ( -r $path )
    {
        if ( ! open JSON_CONFIG, '<', $path )
        {
            status  500;
            message "Error while opening json config file for read : " . $!;
            output  1;
        }

        my $json;
        foreach my $line (<JSON_CONFIG>)
        {
            if ( $line =~ /^([^#;]*)([#;].*)?$/ )
            {
                $json .= $1;
            }
        }
        close JSON_CONFIG;

        eval {
            $config = decode_json( $json );
        };

        if ( $@ )
        {
            status  500;
            message "Error while decoding json config: " . $@;
            output  1;
        }

        if ( ref $config eq "HASH" and JSON::is_bool($config->{'enabled'}) and ! $config->{'enabled'} )
        {
            message "Probe is disabled";
            output  12;
        }
    }
    else
    {
        $config = shift || {};
    }
}

###
# SAVE / LOAD PERSISTANT DATA
###

sub save
{
    return unless $persist;

    my $path = $PERSIST_PATH . "/" . $name . ".wigo";

    if ( open PERSIST, '>', $path )
    {
        eval {
            print PERSIST $json->encode($persist)."\n";
        };
        close PERSIST;

        if ( $@ )
        {
            status 300;
            message "can't serialize persistant data : $@";
            output 1;
        }
    }
    else
    {
        status 300;
        message "can't open persistant data file $path for writing : $!";
        output 1;
    }
}

sub restore
{
    my $path = $PERSIST_PATH . "/" . $name . ".wigo";

    return unless -e $path;

    if ( open PERSIST, '<', $path )
    {
        my @lines  = <PERSIST>;
        close PERSIST;

        chomp @lines;
        my $str = join "\n", @lines;
        return unless $str;

        eval {
            $persist = $json->decode( $str );
        };

        if ( $@ )
        {
            status 300;
            message "can't deserialize persistant data : $@";
            output 1;
        }
    }
    else
    {
        status 300;
        message "can't open persistant data file $path for reading : $!";
        output 1;
    }
}

sub getLevel {
    my $code = shift;

    if ( $code == 100 )
    {
        return 'OK';
    }
    elsif ( $code > 100 and $code < 199 )
    {
        return 'INFO';
    }
    elsif ( $code >= 200 and $code < 300 )
    {
        return 'WARN';
    }
    elsif ( $code >= 300 and $code < 500 )
    {
        return 'CRIT';
    }
    else
    {
        return 'ERROR';
    }
}

1;
