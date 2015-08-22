#!/usr/bin/perl

use strict;
use warnings;

use Cwd qw(abs_path);
use File::Basename qw(dirname);
use lib dirname(abs_path($0)) . '/../../lib';
use if defined $ENV{'WIGO_PROBE_LIB_ROOT'}, lib => $ENV{'WIGO_PROBE_LIB_ROOT'};
use Wigo::Probe qw/:all/;

###
# DEFAULT CONFIG
###

my $conf = {
    'status' => 100,
    'message' => 'dummy',
    'exit' => 0,
    'sleep'  => 0,
    'stderr' => ""
};

init( config => $conf );

message config->{'message'};
raise   config->{'status'};

details->{'foo'} = 'bar';
add_metric { "Tags" => { "foo" => "bar" }, "Value" => 26 };

print STDERR config->{'stderr'} if config->{'stderr'};

sleep config->{'sleep'} if config->{'sleep'};

output  config->{'exit'};