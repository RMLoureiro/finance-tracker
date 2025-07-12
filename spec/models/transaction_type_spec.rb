# frozen_string_literal: true

require "rails_helper"

RSpec.describe TransactionType do
  it "has all columns" do
    expect(subject).to have_db_column(:tittle)
  end
end
