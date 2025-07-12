# frozen_string_literal: true

require "rails_helper"
RSpec.describe Transaction do
  it "has all columns" do
    expect(subject).to have_db_column(:ammount)
  end

  it "has relations" do
    expect(subject).to belong_to(:account)
    expect(subject).to belong_to(:transaction_type)
  end
end
