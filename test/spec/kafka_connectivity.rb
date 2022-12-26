require 'rubygems'
require 'net/ssh'

@hostname = "sshuser@tst-eastus-kafka-hri1.azurehdinsight.net"
@username = "admin"
@password = "Kafka@1234"
@cmd = "/usr/hdp/current/kafka-broker/bin/kafka-topics.sh --list --zookeeper hn0-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net
hn1-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net wn0-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net
wn1-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net wn2-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net
wn3-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net zk0-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net
zk2-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net zk3-tls-hi.jad2revccccubh5rkzzspwwhwe.bx.internal.cloudapp.net"

begin
  puts "SSHing #{@hostname} ..."
  Net::SSH.start( @hostname, @username, :password => @password ) do |ssh|
    puts ssh.exec!('date')
    puts "Logging out..."
  end
  # ssh = Net::SSH.start(@hostname, @username, :password => @password)
  # res = ssh.exec!(@cmd)
  # ssh.close
  #puts res
rescue
  puts "Unable to connect to #{@hostname} using #{@username}/#{@password}"
end
