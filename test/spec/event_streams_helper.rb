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

end