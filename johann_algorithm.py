# idea: make a matrix where p11 is the price per unit of matching sBid1 with bBid1
# as the bids are consumed, reduce the quantity until the line/row is removed.

#     sBid1 sBid2 sBid3
#bBid1 p11   p12    p13
#bBid2 p21   p22    p23
#bBid3 p31   p32    p33

import pprint

sell_bids = [
        {"price_per_credit": 20, "quantity": 10, "linearlocation": 0},
        {"price_per_credit": 50, "quantity": 10, "linearlocation": 10},
        {"price_per_credit": 40, "quantity": 10, "linearlocation": 30},
        ]

buy_bids = [
        {"price_per_credit": 60, "quantity": 10, "linearlocation": -3},
        {"price_per_credit": 40, "quantity": 10, "linearlocation": -4},
        {"price_per_credit": 90, "quantity": 10, "linearlocation": 24},
        ]

def print_matrix(matrix):
    print("------------")
    print("Matrix:")
    for row in matrix:
        for cell in row:
            print([cell[0],cell[1]], end=" ")
        print()
    print("------------")

def multiplier(sell_bid, buy_bid):
    return round(1/abs(sell_bid["linearlocation"] - buy_bid["linearlocation"]),2)

def price_per_credit(sell_bid, buy_bid):
    #price = price_per_credit * quantity
    # quantity_to_burn = min(sell_bid["quantity"], buy_bid["quantity"])*multiplier
    if sell_bid["price_per_credit"] > buy_bid["price_per_credit"]:
        return 0
    average = (sell_bid["price_per_credit"] + buy_bid["price_per_credit"]) / 2
    return round(average,2)

# TODO: THIS IS NOT DONE YET
def johann_algorithm(sell_bids, buy_bids):
    def mount_matrix(sell_bids, buy_bids):
        matrix = []
        for sell_bid in sell_bids:
            row = []
            if sell_bid["quantity"] == 0:
                continue
            for buy_bid in buy_bids:
                if buy_bid["quantity"] == 0:
                    continue
                ppc = price_per_credit(sell_bid, buy_bid)
                mult = multiplier(sell_bid, buy_bid)
                row.append([ppc, mult, buy_bid, sell_bid])
            matrix.append(row)
        return matrix

    def get_combination_with_highest_multiplier(matrix):
        max_mult_combination = [.0, .0]
        for row in matrix:
            for combination in row:
                if combination[1] > max_mult_combination[1]:
                    max_mult_combination = combination
        return max_mult_combination

    def satisfy_bid(bid, matched_bids):
        which_bid = bid[2]["quantity"] - bid[3]["quantity"]
        matched_bids.append({"sell_bid": bid[3], "buy_bid": bid[2], "quantity": min(bid[2]["quantity"], bid[3]["quantity"])})
        if which_bid <= 0:
            bid[2]["quantity"] = 0
        if which_bid >= 0:
            bid[3]["quantity"] = 0

    matrix = mount_matrix(sell_bids, buy_bids)
    matched_bids = []
    while len(matrix) > 0:
        print_matrix(matrix)
        bid = get_combination_with_highest_multiplier(matrix)
        print("-----------------")
        print("satisfying bid: ")
        pprint.pprint(bid)
        print("-----------------")
        satisfy_bid(bid, matched_bids)
        matrix = mount_matrix(sell_bids, buy_bids)

    pprint.pprint(matched_bids)

    # TODO: i can satisfy the bids with higher multiplier first, as they represent a more
    # efficient carbon sinking. Also, the profit of the multiplier must be split between the
    # buyer and the seller.

johann_algorithm(sell_bids, buy_bids)
