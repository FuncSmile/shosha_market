import { Injectable, NotFoundException } from '@nestjs/common';
import { PrismaService } from '../../database/prisma.service';
import { SetPricingDto, BulkPricingDto } from './dto';

@Injectable()
export class PricingService {
  constructor(private prisma: PrismaService) {}

  async setPricing(setPricingDto: SetPricingDto) {
    const { variantId, userCode, price } = setPricingDto;

    // Verify variant exists
    const variant = await this.prisma.productVariant.findUnique({
      where: { id: variantId },
    });

    if (!variant) {
      throw new NotFoundException(`Variant dengan ID ${variantId} tidak ditemukan`);
    }

    // Upsert pricing (update if exists, create if not)
    const pricing = await this.prisma.pricing.upsert({
      where: {
        variantId_userCode: {
          variantId,
          userCode,
        },
      },
      update: { price },
      create: {
        variantId,
        userCode,
        price,
      },
      include: {
        variant: {
          select: {
            id: true,
            name: true,
            sku: true,
            product: {
              select: {
                id: true,
                name: true,
              },
            },
          },
        },
      },
    });

    return pricing;
  }

  async bulkSetPricing(bulkPricingDto: BulkPricingDto) {
    const results: any[] = [];

    for (const pricingData of bulkPricingDto.pricings) {
      const result = await this.setPricing(pricingData);
      results.push(result);
    }

    return {
      message: `${results.length} pricing berhasil diset`,
      data: results,
    };
  }

  async getPricingForVariant(variantId: string) {
    const pricings = await this.prisma.pricing.findMany({
      where: { variantId },
      include: {
        variant: {
          select: {
            id: true,
            name: true,
            sku: true,
            product: {
              select: {
                id: true,
                name: true,
                imageUrl: true,
              },
            },
          },
        },
      },
      orderBy: { userCode: 'asc' },
    });

    return pricings;
  }

  async getPricingForUser(variantId: string, userCode: string) {
    const pricing = await this.prisma.pricing.findUnique({
      where: {
        variantId_userCode: {
          variantId,
          userCode,
        },
      },
      include: {
        variant: {
          select: {
            id: true,
            name: true,
            sku: true,
            stock: true,
            product: {
              select: {
                id: true,
                name: true,
                description: true,
                imageUrl: true,
              },
            },
          },
        },
      },
    });

    if (!pricing) {
      throw new NotFoundException(
        `Pricing untuk variant ${variantId} dengan user code ${userCode} tidak ditemukan`,
      );
    }

    return pricing;
  }

  async getAllPricingsByUserCode(userCode: string) {
    return this.prisma.pricing.findMany({
      where: { userCode },
      include: {
        variant: {
          select: {
            id: true,
            name: true,
            sku: true,
            stock: true,
            product: {
              select: {
                id: true,
                name: true,
                description: true,
                isActive: true,
                imageUrl: true,
              },
            },
          },
        },
      },
      orderBy: {
        variant: {
          product: {
            name: 'asc',
          },
        },
      },
    });
  }

  async deletePricing(variantId: string, userCode: string) {
    const pricing = await this.prisma.pricing.findUnique({
      where: {
        variantId_userCode: {
          variantId,
          userCode,
        },
      },
    });

    if (!pricing) {
      throw new NotFoundException(
        `Pricing untuk variant ${variantId} dengan user code ${userCode} tidak ditemukan`,
      );
    }

    await this.prisma.pricing.delete({
      where: {
        variantId_userCode: {
          variantId,
          userCode,
        },
      },
    });

    return { message: 'Pricing berhasil dihapus' };
  }
}
