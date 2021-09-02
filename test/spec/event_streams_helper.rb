# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class EventStreamsHelper

  def initialize
    @helper = Helper.new
  end

  def get_topics
    @helper.exec_command("bx es topics").split("\n").map { |t| t.strip }
  end

  def create_topic(topic, partitions)
    unless get_topics.include?(topic)
      @helper.exec_command("ibmcloud es topic-create #{topic} -p #{partitions}")
      Logger.new(STDOUT).info("Topic #{topic} created.")
    end
  end

  def delete_topic(topic)
    if get_topics.include?(topic)
      @helper.exec_command("ibmcloud es topic-delete #{topic} -f")
      Logger.new(STDOUT).info("Topic #{topic} deleted.")
    end
  end

end