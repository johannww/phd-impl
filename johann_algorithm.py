# idea: make a matrix where p11 is the price per unit of matching sBid1 with bBid1
# as the bids are consumed, reduce the quantity until the line/row is removed.

#     sBid1 sBid2 sBid3
#bBid1 p11   p12    p13
#bBid2 p21   p22    p23
#bBid3 p31   p32    p33

import pprint

sell_bids = [
        {"price_per_credit": 20, "quantity": 5, "linearlocation": 0},
        {"price_per_credit": 50, "quantity": 10, "linearlocation": 10},
        {"price_per_credit": 40, "quantity": 10, "linearlocation": 30},
        ]

buy_bids = [
        {"price_per_credit": 19.9, "quantity": 10, "linearlocation": -3},
        {"price_per_credit": 40, "quantity": 10, "linearlocation": -4},
        {"price_per_credit": 90, "quantity": 10, "linearlocation": 24},
        ]

debug = True

def print_debug(*args):
    if not debug:
        return
    for arg in args:
        print(arg, end=" ")
    print()

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

def price_per_credit(sell_bid, buy_bid, mult):
    print("---- price per credit ----")
    quantity = min(sell_bid["quantity"], buy_bid["quantity"])
    buyer_total_extra_tokens = seller_total_extra_tokens =  quantity * mult/2
    print_debug("total extra tokens: ", quantity*mult)
    print_debug("buyer's extra tokens: ", buyer_total_extra_tokens)
    print_debug("seller's extra tokens: ", seller_total_extra_tokens)
    acquirable_quantity = quantity + buyer_total_extra_tokens + seller_total_extra_tokens
    print_debug("total acquirable quantity: ", acquirable_quantity)

    to_be_aquired = 0
    if buy_bid["quantity"] >= acquirable_quantity:
        to_be_aquired = buy_bid["quantity"]
    else:
        to_be_aquired = buy_bid["quantity"]/(1+mult) # buy_bid["quantity"] = to_be_aquired + to_be_aquired*mult
    print_debug("actual quantity buyer needs from sell bid: ", to_be_aquired)

    # buyer will pay for the credits and proportionally to the extra seller tokens
    buyer_is_willing_to_pay_total = buy_bid["price_per_credit"]*(to_be_aquired+to_be_aquired*mult/2)
    print_debug("buyer is willing to pay total: ", buyer_is_willing_to_pay_total)
    print_debug("buyer willing price per credit: ", buyer_is_willing_to_pay_total/to_be_aquired)
    buyer_is_willing_to_pay_per_credit_from_seller_perspective = buyer_is_willing_to_pay_total/to_be_aquired
    print_debug("seller desired price per credit: ", sell_bid["price_per_credit"])

    if sell_bid["price_per_credit"] > buyer_is_willing_to_pay_per_credit_from_seller_perspective:
        return 0
    average = (sell_bid["price_per_credit"] + buyer_is_willing_to_pay_per_credit_from_seller_perspective) / 2
    print_debug("cutting price to ", average)
    print_debug("buyer effectively pays ", average*(to_be_aquired))
    print_debug("buyer effectively pays per credit ", average)
    print_debug("seller effectively receives ", average*to_be_aquired)
    print_debug("seller effectively receives per credit ", average)
    return {"cutting_price": round(average,2), "quantity": to_be_aquired}

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
                mult = multiplier(sell_bid, buy_bid)
                ppc = price_per_credit(sell_bid, buy_bid, mult)
                if ppc == 0:
                    continue
                row.append([ppc, mult, buy_bid, sell_bid])
            matrix.append(row)
        return matrix

    def get_combination_with_highest_multiplier(matrix):
        max_mult_combination = [.0, .0]
        for row in matrix:
            for combination in row:
                if combination[1] > max_mult_combination[1]:
                    max_mult_combination = combination
        if max_mult_combination[1] == 0:
            return None
        return max_mult_combination

    def satisfy_bid(bid, matched_bids):
        if bid == None:
            return
        quantity = min(bid[2]["quantity"], bid[3]["quantity"])
        matched_bids.append({"sell_bid": bid[3], "buy_bid": bid[2], "quantity": min(bid[2]["quantity"], bid[3]["quantity"])})
        bid[2]["quantity"] -= quantity
        bid[3]["quantity"] -= quantity

    matrix = mount_matrix(sell_bids, buy_bids)
    matched_bids = []
    len_matched_bids = 0
    while True:
        print_matrix(matrix)
        bid = get_combination_with_highest_multiplier(matrix)
        if bid == None:
            break
        print("-----------------")
        print("satisfying bid: ")
        pprint.pprint(bid)
        print("-----------------")
        satisfy_bid(bid, matched_bids)
        matrix = mount_matrix(sell_bids, buy_bids)

    pprint.pprint(matched_bids)

    print("Bids not fully matched")
    buy_and_sell_bids = sell_bids + buy_bids
    pprint.pprint([x for x in buy_and_sell_bids if x["quantity"] > 0])

    # TODO: i can satisfy the bids with higher multiplier first, as they represent a more
    # efficient carbon sinking. Also, the profit of the multiplier must be split between the
    # buyer and the seller.

johann_algorithm(sell_bids, buy_bids)
